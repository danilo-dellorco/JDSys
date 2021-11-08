package impl

import (
	"fmt"
	"log"
	"net/rpc"
	"progetto-sdcc/utils"
)

type Args0 struct{}
type Args1 struct {
	Key string
}
type Args2 struct {
	Key   string
	Value string
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

func PrintMethodList() {
	fmt.Println("=============== METHODS LIST ===============")
	fmt.Println("1) Get")
	fmt.Println("2) Put")
	fmt.Println("3) Update")
	fmt.Println("4) Delete")
	fmt.Println("5) List")
	fmt.Println("============================================")
}

func Get() {
	var key string
	fmt.Print("Insert the Desired Key: ")
	fmt.Scanln(&key)
	fmt.Println(key)
	testGetRPC(key)
}

/*
Funzione di Debug utile per testare le RPC in locale
*/
func testGetRPC(key string) {
	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	args := Args1{}
	args.Key = key
	fmt.Println(key)
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
	args.Key = "Key_PutRPC"
	args.Value = "Value_PutRPC"
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
	args.Key = "Key_PutRPC"
	args.Value = "NewValue_PutRPC"
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
	args.Key = "TestKey"
	var reply string
	err = client.Call("RPCservice.DeleteRPC", args, &reply)
	if err != nil {
		log.Fatal("GetRPC error:", err)
	}
	fmt.Println("Risposta RPC:", reply)
}
