package main

import (
	"os"
	"../main/conn"
	"fmt"
)

func main() {
	var mode string
	var msg string = "Server Listening For Connection:"
	mode = os.Args[1]

	fmt.Println(msg,mode)
	
	if mode == "tcp"{
		conn.ListenTcpConnection()
	}

	if mode == "http"{
		conn.ListenHttpConnection()
	}
}
