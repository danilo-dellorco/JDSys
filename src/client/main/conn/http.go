package conn

import (
	"fmt"
	"log"
	"net/rpc"
)

type Args1 struct {
	A, B int
}

func HttpConnect(serverAddress string) {
	client, err := rpc.DialHTTP("tcp", serverAddress+":1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	// Synchronous call
	args := Args1{7, 8}
	var reply int
	err = client.Call("Arith.Multiply", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Printf("Moltiplicazione: %d*%d=%d\n", args.A, args.B, reply)
}
