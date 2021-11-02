package main

import (
	"fmt"
	"os"

	"aws"
	"aws/awserr"
	"aws/session"
	"service/elb"
)

func main() {
	var serverAddress string
	serverAddress = os.Args[1]
	if len(os.Args) != 2 {
		fmt.Printf("Usage: go run client.go SERVER_IP\n")
	}
	impl.GetMethodsList(serverAddress)
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

func ExampleELB_DescribeInstanceHealth_shared00() {
	svc := elb.New(session.New())
	input := &elb.DescribeInstanceHealthInput{
		LoadBalancerName: aws.String("my-load-balancer"),
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
