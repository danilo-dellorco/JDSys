package utils

import (
	"net/rpc"
	"time"
)

/*
Permette di instaurare una connessione HTTP con il server all'indirizzo e porta specificati.
Utilizzato per connettersi al Load Balancer
*/
func HttpConnect(addr string, port string) (*rpc.Client, error) {
retry:
	client, err := rpc.DialHTTP("tcp", addr+port)
	if err != nil {
		time.Sleep(DIAL_RETRY)
		goto retry
	}
	return client, err
}
