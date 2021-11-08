package impl

import (
	"fmt"
	"log"
	"net/rpc"
	"progetto-sdcc/utils"
)

type EmptyArguments struct{}

/*
Instaura una connessione HTTP con il Load Balancer, specificando in input il suo url
*/
func HttpConnect(serverAddress string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", serverAddress+utils.RPC_PORT)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return client, err
}

/*
RPC che permette di ottenere la lista dei metodi disponibili sul nodo remoto contattato
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
