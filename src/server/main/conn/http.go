package conn

import (
	"log"
	"net/http"
	//"net/rpc"
	//"fmt"
)

func ListenHttpConnection() {
	//service := new(ServizioDiProva)
	//rpc.Register(service)

	//rpc.HandleHTTP()

	e := http.ListenAndServe(":1234", nil)
	if e != nil {
		log.Fatal("Listen error: ", e)
	}
}
