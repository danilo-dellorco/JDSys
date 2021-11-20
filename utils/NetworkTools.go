package utils

import (
	"log"
	"net/rpc"
)

/*
Permette di instaurare una connessione HTTP con il server all'indirizzo e porta specificati.
*/
func HttpConnect(addr string, port string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", addr+port)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return client, err
}
