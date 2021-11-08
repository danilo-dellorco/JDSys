package impl

import (
	"fmt"
	"log"
	"net/rpc"
	"progetto-sdcc/utils"
)

type Args0 struct{}
type Args2 struct {
	key   string
	value string
}

/*
Instaura una connessione HTTP con il Load Balancer, specificando in input il suo url
*/
func HttpConnect(serverAddress string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", serverAddress+utils.RPC_PORT)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return client, err
}

/*
RPC che permette di ottenere la lista dei metodi disponibili sul nodo remoto contattato
*/
func GetMethodsList(serverAddress string) {
	args := Args0{}
	var reply string

	client, err := HttpConnect(serverAddress)
	err = client.Call("ServizioDiProva.ListMethods", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Printf(reply)
}

/*
Funzione di Debug utile per testare le RPC in locale
*/
func testGetRPC() {
	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	args := Args2{}
	args.key = "TestKeyErr"
	var reply string
	err = client.Call("RPCservice.GetRPC", args, &reply)
	if err != nil {
		log.Fatal("GetRPC error:", err)
	}
	fmt.Println("Risposta RPC:", reply)
}

/*
Funzione di Debug utile per testare le RPC in locale. Sarà identico a come il client dovrà invocare Get e Put
*/
func testPutRPC() {
	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	args := Args2{}
	args.key = "Key_PutRPC"
	args.value = "Value_PutRPC"
	var reply string
	err = client.Call("RPCservice.PutRPC", args, &reply)
	if err != nil {
		log.Fatal("GetRPC error:", err)
	}
	fmt.Println("Risposta RPC:", reply)
}

/*
Funzione di Debug utile per testare le RPC in locale. Sarà identico a come il client dovrà invocare Get e Put
*/
func testUpdateRPC() {
	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	args := Args2{}
	args.key = "Key_PutRPC"
	args.value = "NewValue_PutRPC"
	var reply string
	fmt.Println("UpdatingRPC")
	err = client.Call("RPCservice.UpdateRPC", args, &reply)
	if err != nil {
		log.Fatal("GetRPC error:", err)
	}
	fmt.Println("Risposta RPC:", reply)
}

/*
Funzione di Debug utile per testare le RPC in locale
*/
func testDeleteRPC() {

	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	args := Args2{}
	args.key = "TestKey"
	var reply string
	err = client.Call("RPCservice.DeleteRPC", args, &reply)
	if err != nil {
		log.Fatal("GetRPC error:", err)
	}
	fmt.Println("Risposta RPC:", reply)
}
