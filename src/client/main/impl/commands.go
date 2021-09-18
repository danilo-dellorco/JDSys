package impl

import (
	"fmt"
	"log"
	"net/rpc"
)

type EmptyArguments struct {}

func HttpConnect(serverAddress string) (*rpc.Client, error){
	client, err := rpc.DialHTTP("tcp", serverAddress+":1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return client,err
}

func GetMethodsList(serverAddress string){
	args:= EmptyArguments{}
	var reply string

	client, err:=HttpConnect(serverAddress)
	err = client.Call("ServizioDiProva.ListMethods", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Printf(reply)
}
