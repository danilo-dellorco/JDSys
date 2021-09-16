package main

import (
	"../main/conn"
	"os"
	"fmt"
)

func main() {
	var addr string

	addr = os.Args[1]

	if len(os.Args)!=2 {
		fmt.Printf("Usage: go run client.go SERVER_IP\n")
	}
	conn.HttpConnect(addr)
}