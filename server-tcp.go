package main

import (
	"log"
	"net"
	"net/rpc"
	"os"
)

//struttura per il passaggio dei parametri nella RPC
type Args struct {
	A, B int
}

//"interfaccia" che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
//ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici!
type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func main() {
	arith := new(Arith)
	rpc.Register(arith)

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listen error: ", err)
		os.Exit(1)
	}

	rpc.Accept(listener)
}
