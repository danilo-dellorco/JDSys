package impl

import (
	"errors"
	"fmt"
	"net/rpc"
	"progetto-sdcc/utils"
	"strconv"
	"time"
)

var GET string = "Node.GetRPC"
var PUT string = "Node.PutRPC"
var DEL string = "Node.DeleteRPC"
var APP string = "Node.AppendRPC"

/*
Struttura che mantiene i parametri delle RPC
*/
type Args struct {
	Key     string
	Value   string
	Handler string
	Deleted bool
}

/*
Effettua la RPC per la GET
*/
func GetRPC(key string) {
	args := Args{}
	args.Key = key

	var reply *string

	c := make(chan error)

	client, _ := utils.HttpConnect(utils.LB_DNS_NAME, utils.RPC_PORT)
	defer client.Close()
	go CallRPC(GET, client, args, reply, c)
	rr1_timeout(GET, client, args, reply, c)
}

/*
Effettua la RPC per il PUT
*/
func PutRPC(key string, value string) {
	args := Args{}
	args.Key = key
	args.Value = value

	var reply *string

	c := make(chan error)

	client, _ := utils.HttpConnect(utils.LB_DNS_NAME, utils.RPC_PORT)
	defer client.Close()
	go CallRPC(PUT, client, args, reply, c)
	rr1_timeout(PUT, client, args, reply, c)
}

/*
Effettua la RPC per l'APPEND
*/
func AppendRPC(key string, value string) {
	args := Args{}
	args.Key = key
	args.Value = value

	var reply *string

	c := make(chan error)

	client, _ := utils.HttpConnect(utils.LB_DNS_NAME, utils.RPC_PORT)
	defer client.Close()
	go CallRPC(APP, client, args, reply, c)
	rr1_timeout(APP, client, args, reply, c)
}

/*
Effettua la RPC per il DELETE
*/
func DeleteRPC(key string) {
	args := Args{}
	args.Key = key

	var reply *string

	c := make(chan error)

	client, _ := utils.HttpConnect(utils.LB_DNS_NAME, utils.RPC_PORT)
	defer client.Close()
	go CallRPC(DEL, client, args, reply, c)
	rr1_timeout(DEL, client, args, reply, c)
}

/*
Goroutine per l'implementazione della semantica at-least-once.
Vengono effettuate fino a RR1_RETRIES ritrasmissioni, altrimenti si assume che il server sia crashato.
*/
func rr1_timeout(rpc string, client *rpc.Client, args Args, reply *string, c chan error) {
	signal := make(chan bool)
	res := errors.New("Timeout")
	check := 0
restart_timer:
	for i := 0; i < utils.RR1_RETRIES; i++ {
		go check_timeout(signal)
		select {
		// Scade timer per la ritrasmissione
		case <-signal:
			check++
			utils.PrintTs("Timeout elapsed, send new request n°" + strconv.Itoa(check) + "...")
			go CallRPC(rpc, client, args, reply, c)

		// Arriva risposta dal server
		case res = <-c:
			if res.Error() == "Success" {
				break restart_timer
			}
		}
	}
	// Effettuate tutte le ritrasmissioni possibili e non si riceve alcuna risposta
	if check == utils.RR1_RETRIES && res.Error() != "Success" {
		utils.PrintTs("Server unreachable!")
	}
}

/*
Effettua una generica chiamata RPC, gestendo anche il meccanismo RR1
*/
func CallRPC(rpc string, client *rpc.Client, args Args, reply *string, c chan error) {
	err := client.Call(rpc, args, &reply)
	defer client.Close()
	if err != nil {
		c <- err
		utils.PrintTs("RPC error " + err.Error())

	} else {
		c <- errors.New("Success")

		fmt.Println(*reply)
	}
	return
}

/*
Controlla lo scadere del timeout
*/
func check_timeout(check chan bool) {
	time.Sleep(utils.RR1_TIMEOUT)
	check <- true
}

/****************************************************************
* Testing RPC Call                                              *
****************************************************************/
/*
Effettua una richiesta di Put, una di Update, una di Get, una di Append e una di Delete, misurando poi il tempo medio di risposta
*/
func MeasureResponseTime() {
	utils.PrintHeaderL2("Starting Measuring Response Time")
	rt1 := MeasurePut("rt_key", "rt_value")
	rt2 := MeasurePut("rt_key", "rt_value_upd")
	rt3 := MeasureGet("rt_key")
	rt4 := MeasureAppend("rt_key", "rt_value_app")
	rt5 := MeasureDelete("rt_key")

	total := rt1 + rt2 + rt3 + rt4 + rt5
	meanRt := total / 5
	utils.PrintLineL2()
	fmt.Println("Mean Response Time:", meanRt)
	utils.PrintLineL2()
	EnterToContinue()
}

/*
Permette al client di recuperare il valore associato ad una precisa chiave contattando il LB
*/
func MeasureGet(key string) time.Duration {
	utils.PrintTs("Measuring Get...")
	start := utils.GetTimestamp()
	GetRPC(key)
	end := utils.GetTimestamp()
	ts := end.Sub(start)
	utils.PrintTs(fmt.Sprintf("Get Executed in %d", ts))
	return ts
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func MeasurePut(key string, value string) time.Duration {
	utils.PrintTs("Measuring Put...")

	start := utils.GetTimestamp()
	PutRPC(key, value)
	end := utils.GetTimestamp()
	ts := end.Sub(start)
	utils.PrintTs(fmt.Sprintf("Put Executed in %d", ts))
	return ts
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func MeasureAppend(key string, value string) time.Duration {
	utils.PrintTs("Measuring Append...")

	start := utils.GetTimestamp()
	AppendRPC(key, value)
	end := utils.GetTimestamp()

	ts := end.Sub(start)
	utils.PrintTs(fmt.Sprintf("Append Executed in %d", ts))
	return ts
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func MeasureDelete(key string) time.Duration {
	utils.PrintTs("Measuring Delete...")
	start := utils.GetTimestamp()
	DeleteRPC(key)
	end := utils.GetTimestamp()
	ts := end.Sub(start)
	utils.PrintTs(fmt.Sprintf("Delete Executed in %d", ts))
	return ts
}
