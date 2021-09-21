package main

import (
	"fmt"
	"os"
	"progetto-sdcc/client/impl"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elb"
)

func main() {
	var serverAddress string
	serverAddress = os.Args[1]
	if len(os.Args) != 2 {
		fmt.Printf("Usage: go run client.go SERVER_IP\n")
	}
	//impl.GetMethodsList(serverAddress)
	fmt.Println("QUI1")
	//stampaAllarmi()
	ExampleELB_DescribeInstanceHealth_shared00()
	fmt.Println("QUI1")
	for {
		var cmd string
		fmt.Printf("Inserisci un comando: ")
		fmt.Scanln(&cmd)

		switch cmd {
		case "list":
			impl.GetMethodsList(serverAddress)
		}
	}
}

func createSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")})
	fmt.Println(err)
	return sess
}

func stampaAllarmi() {
	sess := createSession()
	svc := cloudwatch.New(sess)
	resp, err := svc.DescribeAlarms(nil)
	for _, alarm := range resp.MetricAlarms {
		fmt.Println(*alarm.AlarmName)
	}
	fmt.Println(err)
}

func ExampleELB_DescribeInstanceHealth_shared00() {
	sess := createSession()

	svc := elb.New(sess)
	input := &elb.DescribeInstanceHealthInput{
		LoadBalancerName: aws.String("ProvaLoadBalancer"),
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
