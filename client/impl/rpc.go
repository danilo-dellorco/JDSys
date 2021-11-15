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

	c := make(chan error)

	client, _ := HttpConnect()
	go rr1_timeout(client, args, reply, c)
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

func CallRPC(client *rpc.Client, args Args1, reply *string, c chan error) {
	go func() { c <- client.Call("RPCservice.GetRPC", args, &reply) }()
	select {
	case err := <-c:
		if err != nil {
			log.Fatal("RPC error: ", err)
		} else {
			fmt.Println(*reply)
		}

	case <-time.After(15 * time.Second):
		return
	}
}

func rr1_timeout(client *rpc.Client, args Args1, reply *string, c chan error) {
	//ciclo che deve essere fatto tante volte quante vogliamo ritrasmettere
	for {
		time.Sleep(utils.RR1_TIMEOUT)
		fmt.Println("scaduto timer")
		res := <-c
		fmt.Println(res)
		//errore, riprovo
		if res != nil {
			CallRPC(client, args, reply, c)
		} else {
			break
		}
	}
}
