# Auto DDNS Security Lambda

Update AWS Security Group rules to an IP resolved from a DNS hostname. Useful to dynamically allow ingress from a DDNS
hostname.

## How it Works

Every time the Lambda is invoked it will resolve the configured hostname to a single IPv4 address.

It will inspect each of the configured security group IDs.

It updates any of the ingress or egress rules that contain `@DDNS` in the description, is a IPv4 rule and does not match
the resolved IPv4 address with the new target IPv4 address.

## Setup

### Prerequisites

- Configure a DDNS hostname on your router
- Create one or more security groups containing ingress or egress rules which you would like to allow access from your
  IP
    - Somewhere in the rule description enter `@DDNS`

### Lambda Setup

- Compile and package the Lambda function using
  ```
  bash package.sh
  ```
- Create a new IAM role
    - Attach the `AWSLambdaBasicExecutionRole` policy
    - Create a new inline policy using [ManageSecurityGroups.json]()
- Create a new function in the Lambda console
    - Use the `go1.x` runtime
    - Use the IAM role created in the previous step as the execution role
- Once created upload the `dist/functions.zip` archive
- Change the handler to `main`
- Edit the general configuration and set the memory to `128MB` (the minimum allowed)
- Edit the environment variables to include
    - `SECURITY_GROUP_IDS` A comma delimited list of target security group rules
    - `TARGET_HOSTNAME` Your DDNS hostname

### CloudWatch Schedule

Now the Lambda is configured, you can create a new CloudWatch schedule to check and update the rules at any given
interval.

- Go to the CloudWatch console
- Go to Rules and click `Create rule`
- Select `Schedule` and enter your desired schedule
- Add a Target
    - Select `Lambda function`
    - Select your Lambda function
    - Select `Constant (JSON Text)` as input and enter `{}`

## Limitations

- Only IPv4 addresses are supported


## Command Line Utility

A command line variant is also provided which does the same thing as the Lambda

```
go run github.com/relvacode/lambda-ddns/cmd/ddns -region AWS_REGION -hostname TARGET_HOSTNAME SECURITY_GROUP_ID [SECURITY_GROUP_ID...]
```
