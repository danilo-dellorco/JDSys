package main

import (
	"fmt"
	"os"
	"progetto-sdcc/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

var ELB_ARN_D string = "arn:aws:elasticloadbalancing:us-east-1:427788101608:loadbalancer/net/NetworkLB/8d7f674bf6bc6f73"
var ELB_ARN_J string = "arn:aws:elasticloadbalancing:us-east-1:806961903927:loadbalancer/net/progetto-sdcc-lb/639e06d499fd2aba"
var ELB string

/**
* Struttura contenente tutte le informazioni riguardanti un nodo
**/
type Instance struct {
	ID, PrivateIP, PublicIP string
}

//TODO mettere getActiveNodes() in una goroutine periodica, ad esempio ogni minuto
func main() {
	getTerminatingScalingActivities(getScalingActivities())

	/*
		if len(os.Args) < 2 {
			fmt.Println("Wrong usage: Specify user \"d\" or \"j\"")
			return
		}
		setupUser()
		activeNodes := getActiveNodes()
		fmt.Println(activeNodes)
		return
	*/
}

/**
* Crea una sessione client AWS
**/
func createSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")})
	if err != nil {
		fmt.Println(err)
	}
	return sess
}

/**
* Imposta l'utente corretto della sessione AWS
**/
func setupUser() {
	user := os.Args[1]
	if user == "d" {
		ELB = ELB_ARN_D
	} else {
		ELB = ELB_ARN_J
	}
}

/**
* Ottiene tutte le informazioni relative al Target Group specificato
**/
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
			fmt.Println(err.Error())
		}
	}
	return result
}

/**
* Ottiene lo stato delle istanze collegate al Target Group specificato
**/
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
			fmt.Println(err.Error())
		}
	}
	return result
}

/**
* Ottiene gli ID associati a tutte le istanze healthy
**/
func getHealthyInstancesId(targetHealth *elbv2.DescribeTargetHealthOutput) []string {
	var healthyNodes []string
	descriptions := targetHealth.TargetHealthDescriptions
	for i := 0; i < len(descriptions); i++ {
		actual := descriptions[i].String()
		id := utils.GetStringInBetween(actual, "Id: \"", "\",")
		state := utils.GetStringInBetween(actual, "State: \"", "\"")
		if state == "healthy" {
			healthyNodes = append(healthyNodes, id)
		}
	}
	return healthyNodes
}

/**
* Ottiene tutte le informazioni di una istanza EC2 tramite il suo ID
**/
func getInstanceInfo(instanceId string) *ec2.DescribeInstancesOutput {
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
			fmt.Println(err.Error())
		}
	}
	return result
}

/**
* Ottiene ID, Indirizzo Pubblico e Indirizzo Privato di una istanza EC2
**/
func getInstanceAddress(instanceInfo *ec2.DescribeInstancesOutput) Instance {
	descriptions := instanceInfo.Reservations
	actual := descriptions[0].String()
	id := utils.GetStringInBetween(actual, "InstanceId: \"", "\",")
	public := utils.GetStringInBetween(actual, "PublicIpAddress: \"", "\"")
	private := utils.GetStringInBetween(actual, "PrivateIpAddress: \"", "\"")
	return Instance{id, public, private}
}

/**
* Ritorna gli indirizzi IP di tutti i nodi connessi al load balancer
**/
func getActiveNodes() map[int]Instance {
	nodes := make(map[int]Instance)
	targetGroup := getTargetGroup(ELB)
	targetGroupArn := utils.GetStringInBetween(targetGroup.String(), "TargetGroupArn: \"", "\",")
	targetsHealth := getTargetsHealth(targetGroupArn)
	healthyInstancesList := getHealthyInstancesId(targetsHealth)
	fmt.Println("Healthy Instances: ")
	fmt.Println(healthyInstancesList)

	for i := 0; i < len(healthyInstancesList); i++ {
		instance := getInstanceInfo(healthyInstancesList[i])
		nodes[i] = getInstanceAddress(instance)
	}

	fmt.Println("Address Healthy Instances: ")
	for key, element := range nodes {
		fmt.Println("Key: ", key, "=>", "Element:", element)
	}

	return nodes
}

// Questa funzione andrà utilizzata dal nodo
func getScalingActivities() *autoscaling.DescribeScalingActivitiesOutput {
	sess := createSession()
	svc := autoscaling.New(sess)
	input := &autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String("SDCC-autoscaling"),
	}

	result, err := svc.DescribeScalingActivities(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case autoscaling.ErrCodeInvalidNextToken:
				fmt.Println(autoscaling.ErrCodeInvalidNextToken, aerr.Error())
			case autoscaling.ErrCodeResourceContentionFault:
				fmt.Println(autoscaling.ErrCodeResourceContentionFault, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	return result
}

func getTerminatingScalingActivities(activityList *autoscaling.DescribeScalingActivitiesOutput) []autoscaling.Activity {
	var terminatingNodes []string
	activities := activityList.Activities
	TERMINATING_START := "Description: \"Terminating EC2 instance:"
	TERMINATING_END := " -"

	for i := 0; i < len(activities); i++ {
		actual := activities[i].String()
		fmt.Println(actual)
		progress := utils.GetStringInBetween(actual, "Progress: ", ",")
		if progress != "100" {
			status := utils.GetStringInBetween(actual, "StatusCode: \"", "\"\n")
			if status == "WaitingForELBConnectionDraining" || status == "InProgress" {
				nodeId := utils.GetStringInBetween(actual, TERMINATING_START, TERMINATING_END)
				terminatingNodes = append(terminatingNodes, nodeId)
				fmt.Println("Status: ", status)
				fmt.Println("nodeId: ", nodeId)
				terminatingNodes = append(terminatingNodes, nodeId)
			}
		}
	}
	return nil
}
