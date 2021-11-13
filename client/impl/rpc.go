package impl

import (
	"fmt"
	"log"
)

func GetRPC(key string) {
	args := Args1{}
	args.Key = key

	var reply *string

	client, _ := HttpConnect()
	err := client.Call("RPCservice.GetRPC", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
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
