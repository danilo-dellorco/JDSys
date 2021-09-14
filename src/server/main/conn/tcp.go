package conn

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"fmt"
)

//struttura per il passaggio dei parametri nella RPC
type Args1 struct {
	A, B int
}

//"interfaccia" che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
//ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici!
type Arith1 int

func (t *Arith1) Multiply(args *Args1, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func ListenTcpConnection() {
	arith := new(Arith1)
	rpc.Register(arith)

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listen error: ", err)
		os.Exit(1)
	}

	rpc.Accept(listener)
	fmt.Printf("Connessione stabilita con il client")
	
}
