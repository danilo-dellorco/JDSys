package utils

import (
	"fmt"
	"net/rpc"
)

/*
Permette di instaurare una connessione HTTP con il server all'indirizzo e porta specificati.
Utilizzato per connettersi al Load Balancer
*/
func HttpConnect(addr string, port string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", addr+port)
	if err != nil {
		fmt.Println("Connection error: " + err.Error())
	}
	return client, err
}
