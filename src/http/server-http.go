package http

import (
	"log"
	"net/http"
	"net/rpc"
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

	rpc.HandleHTTP()

	e := http.ListenAndServe(":1234", nil)
	if e != nil {
		log.Fatal("Listen error: ", e)
	}
}
