package conn

import (
	"fmt"
	"log"
	"net/rpc"
)

type Args2 struct {
	A, B int
}

func TcpConnect(serverAddress string) {
	client, err := rpc.Dial("tcp", serverAddress+":1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	// Synchronous call
	args := Args2{7, 8}
	var reply int
	err = client.Call("Arith1.Multiply", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Printf("Moltiplicazione: %d*%d=%d\n", args.A, args.B, reply)
}
