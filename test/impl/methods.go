package impl

import (
	"fmt"
	"net/rpc"
	"progetto-sdcc/utils"
	"time"
)

var WORKLOAD_GET []int
var WORKLOAD_PUT []int
var WORKLOAD_APP []int

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
Parametri per le operazioni di Get e Delete
*/
type Args1 struct {
	Key string
}

/*
Parametri per le operazioni di Put e Update
*/
type Args2 struct {
	Key   string
	Value string
}

/*
Permette al client di recuperare il valore associato ad una precisa chiave contattando il LB
*/
func TestGet(key string, print bool, id int) time.Duration {
	WORKLOAD_GET[id] = 1

	start := utils.GetTimestamp()
	GetRPC(key, print)
	WORKLOAD_GET[id] = 0
	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func TestPut(key string, value string, print bool, id int) time.Duration {
	WORKLOAD_PUT[id] = 1

	start := utils.GetTimestamp()
	PutRPC(key, value, print)
	WORKLOAD_PUT[id] = 0
	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func TestAppend(key string, value string, print bool, id int) time.Duration {
	WORKLOAD_APP[id] = 1

	start := utils.GetTimestamp()
	AppendRPC(key, value, print)
	WORKLOAD_APP[id] = 0
	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func TestDelete(key string, print bool) time.Duration {
	start := utils.GetTimestamp()
	DeleteRPC(key, print)
	end := utils.GetTimestamp()
	return end.Sub(start)
}

/*
Effettua la RPC per la GET
*/
func GetRPC(key string, print bool) {
	args := Args{}
	args.Key = key

	var reply *string

	client, _ := rpc.DialHTTP("tcp", utils.LB_DNS_NAME+utils.RPC_PORT)
	if client != nil {
		client.Call(GET, args, &reply)
		client.Close()
	}
}

/*
Effettua la RPC per il PUT
*/
func PutRPC(key string, value string, print bool) {
	args := Args{}
	args.Key = key
	args.Value = value

	var reply *string

	client, _ := rpc.DialHTTP("tcp", utils.LB_DNS_NAME+utils.RPC_PORT)
	if client != nil {
		client.Call(PUT, args, &reply)
		client.Close()
	}
}

/*
Effettua la RPC per l'APPEND
*/
func AppendRPC(key string, value string, print bool) {
	args := Args{}
	args.Key = key
	args.Value = value

	var reply *string
	client, _ := rpc.DialHTTP("tcp", utils.LB_DNS_NAME+utils.RPC_PORT)
	if client != nil {
		client.Call(APP, args, &reply)
		client.Close()
	}
}

/*
Effettua la RPC per il DELETE
*/
func DeleteRPC(key string, print bool) {
	args := Args{}
	args.Key = key

	var reply *string
	client, _ := rpc.DialHTTP("tcp", utils.LB_DNS_NAME+utils.RPC_PORT)
	if client != nil {
		client.Call(DEL, args, &reply)
		client.Close()
	}
}

/*
Effettua una richiesta di Put, una di Update, una di Get, una di Append e una di Delete, misurando poi il tempo medio di risposta
*/
func MeasureResponseTime() {
	WORKLOAD_GET = make([]int, 1)
	WORKLOAD_PUT = make([]int, 1)
	WORKLOAD_APP = make([]int, 1)
	utils.PrintHeaderL2("Starting Measuring Response Time")
	utils.PrintTs("Put")
	rt1 := TestPut("rt_key", "rt_value", true, 0)
	rt2 := TestPut("rt_key", "rt_value_upd", true, 0)
	rt3 := TestGet("rt_key", true, 0)
	rt4 := TestAppend("rt_key", "rt_value_app", true, 0)
	rt5 := TestDelete("rt_key", true)

	total := rt1 + rt2 + rt3 + rt4 + rt5
	meanRt := total / 5
	fmt.Println("Mean Response Time:", meanRt)
}
