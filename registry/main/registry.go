package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

type TargetGroup struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	Test []interface{} `json:"test"`
}

//TODO marshal del JSON errore primo parametro di unmarshal
func main() {
	session := createSession()
	fmt.Println(session)
	targetGroupJson := getTargetGroup()
	var targetGroup TargetGroup
	json.Unmarshal(targetGroupJson, &targetGroup)

}

func createSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")})
	fmt.Println(err)
	return sess
}

func getTargetsHealth() {
	sess := createSession()
	svc := elbv2.New(sess)
	input := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:427788101608:targetgroup/TargetNetworkLB/e94c2d0b4c023ec9"),
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
		return
	}

	fmt.Println(result)
}

func getTargetGroup() *elbv2.DescribeTargetGroupsOutput {
	sess := createSession()
	svc := elbv2.New(sess)
	input := &elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:427788101608:loadbalancer/net/NetworkLB/8d7f674bf6bc6f73"),
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
		return result
	}
}

func ExampleELB_DescribeInstanceHealth_shared00() {
	sess := createSession()

	svc := elb.New(sess)
	input := &elb.DescribeInstanceHealthInput{
		LoadBalancerName: aws.String("NetworkLB"),
	}

	result, err := svc.DescribeInstanceHealth(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elb.ErrCodeAccessPointNotFoundException:
				fmt.Println(elb.ErrCodeAccessPointNotFoundException, aerr.Error())
			case elb.ErrCodeInvalidEndPointException:
				fmt.Println(elb.ErrCodeInvalidEndPointException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}
