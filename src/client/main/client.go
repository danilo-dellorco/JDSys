package main

import (
	"../main/conn"
	"os"
)

func main() {
	var mode string
	var addr string

	mode = os.Args[1]
	addr = os.Args[2]
	
	if mode == "tcp"{
		conn.TcpConnect(addr)
	}

	if mode == "http"{
		conn.HttpConnect(addr)
	}
}