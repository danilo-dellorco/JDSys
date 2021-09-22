package main

import (
	"fmt"
	"progetto-sdcc/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

var ELB_ARN string = "arn:aws:elasticloadbalancing:us-east-1:427788101608:loadbalancer/net/NetworkLB/8d7f674bf6bc6f73"
var TEST_INSTANCE string = "i-0a7f1097d88fd8d43"

//TODO marshal del JSON errore primo parametro di unmarshal
func main() {
	targetGroup := getTargetGroup(ELB_ARN)
	targetGroupArn := getTargetGroupArn(targetGroup)
	targetsHealth := getTargetsHealth(targetGroupArn)
	healthyInstancesList := getHealthyInstancesId(targetsHealth)
	for _, id := range healthyInstancesList {
		fmt.Println(getInstanceInfo(id))
	}
	//getInstanceInfo("i-0a7f1097d88fd8d43")
	//targetGroupJson := getTargetGroup()
	//var targetGroup TargetGroup
	//json.Unmarshal(targetGroupJson, &targetGroup)
	return
}

func createSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println(err)
	}
	return sess
}

func getInstanceInfo(instanceId string) string {
	sess := createSession()
	svc := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceId),
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return result.GoString()
	}
	return "nil"
}

func getTargetsHealth(targetGroupArn string) *elbv2.DescribeTargetHealthOutput {
	sess := createSession()
	svc := elbv2.New(sess)
	input := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(targetGroupArn),
	}

	result, err := svc.DescribeTargetHealth(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeInvalidTargetException:
				fmt.Println(elbv2.ErrCodeInvalidTargetException, aerr.Error())
			case elbv2.ErrCodeTargetGroupNotFoundException:
				fmt.Println(elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			case elbv2.ErrCodeHealthUnavailableException:
				fmt.Println(elbv2.ErrCodeHealthUnavailableException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	return result
}

func getTargetGroup(elbArn string) *elbv2.DescribeTargetGroupsOutput {
	sess := createSession()
	svc := elbv2.New(sess)
	input := &elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: aws.String(elbArn),
	}

	result, err := svc.DescribeTargetGroups(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elbv2.ErrCodeLoadBalancerNotFoundException:
				fmt.Println(elbv2.ErrCodeLoadBalancerNotFoundException, aerr.Error())
			case elbv2.ErrCodeTargetGroupNotFoundException:
				fmt.Println(elbv2.ErrCodeTargetGroupNotFoundException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	return result
}

func getInstancePublicIp(instanceInfo string) string {
	return utils.GetStringInBetween(instanceInfo, "PublicIpAddress: \"", "\",")
}

func getInstancePrivateIp(instanceInfo string) string {
	return utils.GetStringInBetween(instanceInfo, "PrivateIpAddress: \"", "\",")
}

func getTargetGroupArn(targetGroupResult *elbv2.DescribeTargetGroupsOutput) string {
	return utils.GetStringInBetween(targetGroupResult.String(), "TargetGroupArn: \"", "\",")
}

/*
func getInstancesFromGroup(groupResult *elbv2.DescribeTargetGroupsOutput) {
	var list = strings.Fields(groupResult.TargetGroups[0].String())
}
*/

func getHealthyInstancesId(targetHealth *elbv2.DescribeTargetHealthOutput) []string {
	//retrieve private and public IP addresses associated to every ec2 instance
	var healthyNodes []string
	descriptions := targetHealth.TargetHealthDescriptions
	for i := 0; i < len(descriptions); i++ {
		actual := descriptions[i].String()
		fmt.Println(actual)
		id := utils.GetStringInBetween(actual, "Id: \"", "\",")
		state := utils.GetStringInBetween(actual, "State: \"", "\"")
		fmt.Println(id)
		fmt.Println(state)
		if state == "healthy" {
			healthyNodes = append(healthyNodes, id)
		}
	}
	return healthyNodes
}
