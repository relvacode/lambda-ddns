package ddns

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"net"
	"strings"
)

const (
	// RuleDescriptionContains is the text the Handler will look for in a security group rule description.
	RuleDescriptionContains = "@DDNS"
)

// New creates a new Handler to manage security groups.
func New(securityGroups []string, region, hostname string) (*Handler, error) {
	s, err := session.NewSession(&aws.Config{
		Region: &region,
	})

	if err != nil {
		return nil, err
	}

	return &Handler{
		securityGroups: securityGroups,
		ec2:            ec2.New(s),
		hostname:       hostname,
	}, nil
}

// Handler handles IP rules for a set of managed AWS security groups.
// It uses the system DefaultResolver to resolve a hostname to IP addresses.
type Handler struct {
	securityGroups []string
	ec2            *ec2.EC2
	hostname       string
}

// Resolve resolves a single IPv4 address for the configured hostname.
// It returns a /32 IP range for that address.
func (l *Handler) Resolve(ctx context.Context) (string, error) {
	addrs, err := net.DefaultResolver.LookupIP(ctx, "ip4", l.hostname)
	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("name not resolved")
	}

	return fmt.Sprintf("%s/32", addrs[0].To4().String()), nil
}

// ManageRules is given a security group ID and a target IP range.
// It looks at all existing rules for that security group which contain RuleDescriptionContains within its description.
// If the rule is for a CidrIpv4 range, and that range does not match the target range,
// then it is updated to the new target range.
func (l *Handler) ManageRules(ctx context.Context, securityGroupId string, targetIpv4Range string) error {
	resp, err := l.ec2.DescribeSecurityGroupRulesWithContext(ctx, &ec2.DescribeSecurityGroupRulesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-id"),
				Values: []*string{&securityGroupId},
			},
		},
	})

	if err != nil {
		return err
	}

	var targetRules []*ec2.SecurityGroupRule
	for _, rule := range resp.SecurityGroupRules {
		ok := rule.CidrIpv4 != nil && rule.Description != nil && strings.Contains(*rule.Description, RuleDescriptionContains)
		if !ok {
			continue
		}

		ok = targetIpv4Range == *rule.CidrIpv4
		if !ok {
			targetRules = append(targetRules, rule)
		}

	}

	if len(targetRules) == 0 {
		log.Printf("No rules to update")
		return nil
	}

	var updates = make([]*ec2.SecurityGroupRuleUpdate, 0, len(targetRules))
	for _, rule := range targetRules {
		log.Printf("%s (%d-%d) %s -> %s", *rule.IpProtocol, *rule.FromPort, *rule.ToPort, *rule.CidrIpv4, targetIpv4Range)

		updates = append(updates, &ec2.SecurityGroupRuleUpdate{
			SecurityGroupRuleId: rule.SecurityGroupRuleId,
			SecurityGroupRule: &ec2.SecurityGroupRuleRequest{
				IpProtocol:  rule.IpProtocol,
				FromPort:    rule.FromPort,
				ToPort:      rule.ToPort,
				Description: rule.Description,
				CidrIpv4:    &targetIpv4Range,
			},
		})
	}

	_, err = l.ec2.ModifySecurityGroupRulesWithContext(ctx, &ec2.ModifySecurityGroupRulesInput{
		GroupId:            &securityGroupId,
		SecurityGroupRules: updates,
	})

	if err != nil {
		return err
	}

	return nil
}

// Update updates all configured security groups.
func (l *Handler) Update(ctx context.Context) error {
	target, err := l.Resolve(ctx)
	if err != nil {
		return err
	}

	log.Printf("Target CIDR is %s", target)

	for _, group := range l.securityGroups {
		log.Printf("Managing security group %s", group)
		err = l.ManageRules(ctx, group, target)
		if err != nil {
			return err
		}
	}

	return nil
}
