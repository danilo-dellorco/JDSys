package services

import (
	"fmt"
	"os"
	"progetto-sdcc/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

var ELB_ARN_D string = "arn:aws:elasticloadbalancing:us-east-1:427788101608:loadbalancer/net/NetworkLB/8d7f674bf6bc6f73"
var ELB_ARN_J string = "arn:aws:elasticloadbalancing:us-east-1:786781699181:loadbalancer/net/sdcc-lb/505f5d098d3c2bc3"
var ELB string

/**
* Struttura contenente tutte le informazioni riguardanti un nodo
**/
type Instance struct {
	ID, PrivateIP string
}

/**
* Crea una sessione client AWS
**/
func createSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("/home/jacopo/.aws/credentials", "default")})
	if err != nil {
		fmt.Println(err)
	}
	return sess
}

/**
* Imposta l'utente corretto della sessione AWS
**/
func SetupUser() {
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
func getInstance(instanceInfo *ec2.DescribeInstancesOutput) Instance {
	descriptions := instanceInfo.Reservations
	actual := descriptions[0].String()
	id := utils.GetStringInBetween(actual, "InstanceId: \"", "\",")
	private := utils.GetStringInBetween(actual, "PrivateIpAddress: \"", "\"")
	return Instance{id, private}
}

/**
* Ritorna gli indirizzi IP di tutti i nodi connessi al load balancer
**/
func GetActiveNodes() []Instance {
	var nodes []Instance
	targetGroup := getTargetGroup(ELB)
	//fmt.Println(targetGroup)
	targetGroupArn := utils.GetStringInBetween(targetGroup.String(), "TargetGroupArn: \"", "\",")
	//fmt.Println(targetGroupArn)
	targetsHealth := getTargetsHealth(targetGroupArn)
	//fmt.Println(targetsHealth)
	healthyInstancesList := getHealthyInstancesId(targetsHealth)
	fmt.Println("Healthy Instances: ")
	fmt.Println(healthyInstancesList)

	nodes = make([]Instance, len(healthyInstancesList))
	for i := 0; i < len(healthyInstancesList); i++ {
		instance := getInstanceInfo(healthyInstancesList[i])
		nodes[i] = getInstance(instance)
	}

	//fmt.Println("Address Healthy Instances: ")
	//for key, element := range nodes {
	//	fmt.Println("Key: ", key, "=>", "Element:", element)
	//}
	return nodes
}

/**
* Ottiene dal Load Balancer la lista delle attività schedulate in termini di ScaleIN e ScaleOUT.
**/
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

/**
* Ottiene tutte gli ID di tutte le istanze che sono nello stato di terminazione
**/
func GetTerminatingInstances() []Instance {
	activityList := getScalingActivities()

	var terminatingNodes []Instance
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
				fmt.Println("Status: ", status)
				fmt.Println("nodeId: ", nodeId)
				instanceInfo := getInstanceInfo(nodeId)
				instance := getInstance(instanceInfo)
				terminatingNodes = append(terminatingNodes, instance)
			}
		}
	}
	return terminatingNodes
}
