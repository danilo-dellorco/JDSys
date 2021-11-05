package impl

import (
	"fmt"
	"log"
	"net/rpc"
)

type EmptyArguments struct{}

/*
Instaura una connessione HTTP con il server/nodo specificato in input
*/
func HttpConnect(serverAddress string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", serverAddress+":80")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return client, err
}

/*
RPC che permette di ottenere la lista dei metodi disponibili nel nodo remoto
*/
func GetMethodsList(serverAddress string) {
	args := EmptyArguments{}
	var reply string

	client, err := HttpConnect(serverAddress)
	err = client.Call("ServizioDiProva.ListMethods", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Printf(reply)
}
