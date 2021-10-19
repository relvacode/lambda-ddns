package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/relvacode/lambda-ddns/ddns"
	"log"
	"os"
	"strings"
)

var handler *ddns.Handler

func init() {
	securityGroupIds := strings.Split(os.Getenv("SECURITY_GROUP_IDS"), ",")
	if len(securityGroupIds) == 0 {
		log.Fatalln("No SECURITY_GROUP_IDS specified")
	}

	targetHostname := os.Getenv("TARGET_HOSTNAME")
	if targetHostname == "" {
		log.Fatalln("No TARGET_HOSTNAME specified")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatalln("No AWS_REGION specified")
	}

	var err error
	handler, err = ddns.New(securityGroupIds, awsRegion, targetHostname)
	if err != nil {
		log.Fatalln(err)
	}
}

func HandleLambdaEvent(ctx context.Context) error {
	return handler.Update(ctx)
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
