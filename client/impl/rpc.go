package impl

import (
	"fmt"
	"log"
	"net/rpc"
	"progetto-sdcc/utils"
	"time"
)

func GetRPC(key string) {
	args := Args1{}
	args.Key = key

	var reply *string

	c := make(chan string)

	client, _ := HttpConnect()
	CallRPC(client, args, reply, c)
	fmt.Println("Risposta RPC:", *reply)
}

func PutRPC(key string, value string) {
	args := Args2{}
	args.Key = key
	args.Value = value

	var reply *string

	client, _ := HttpConnect()
	err := client.Call("RPCservice.PutRPC", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Println("Risposta RPC:", *reply)
}

func AppendRPC(key string, value string) {
	args := Args2{}
	args.Key = key
	args.Value = value
	var reply *string

	client, _ := HttpConnect()
	err := client.Call("RPCservice.AppendRPC", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Println("Risposta RPC:", *reply)
}

func DeleteRPC(key string) {
	args := Args1{}
	args.Key = key
	var reply *string

	client, _ := HttpConnect()
	err := client.Call("RPCservice.DeleteRPC", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Println("Risposta RPC:", *reply)
}

func rr1_timeout(client *rpc.Client, args Args1, reply *string, c chan string) {
	for {
		time.Sleep(utils.RR1_TIMEOUT)
		fmt.Println("scaduto timer")
		res := <-c
		fmt.Println(reply)
		if res == "" {
			CallRPC(client, args, &res, c)
		} else {
			break
		}
	}
}

func CallRPC(client *rpc.Client, args Args1, reply *string, c chan string) {
	go rr1_timeout(client, args, reply, c)
	fmt.Println("prima call")
	err := client.Call("RPCservice.GetRPC", args, &reply)
	fmt.Println("dopo call")
	fmt.Println(reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	c <- *reply
}
