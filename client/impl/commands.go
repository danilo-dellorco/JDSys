package impl

import (
	"fmt"
	"log"
	"net/rpc"
	"progetto-sdcc/utils"
)

/*
Parametri per le operazioni di Get e Delete
*/
type Args1 struct {
	Key string
}

/*
Parametri per le operazioni di Put e Update
*/
type Args2 struct {
	Key   string
	Value string
}

/*
Permette di instaurare una connessione HTTP con il LB fornendo il suo nome DNS.
*/
func HttpConnect(lbAddr string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", lbAddr+utils.RPC_PORT)
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
	fmt.Println("============================================")
}

/*
Permette al client di recuperare il valore associato ad una precisa chiave contattando il LB
*/
func Get(lbAddr string) {
	args := Args1{}
	fmt.Print("Insert the Desired Key: ")
	fmt.Scanln(&args.Key)
	fmt.Println(args.Key)

	var reply *string

	client, _ := HttpConnect(lbAddr)
	err := client.Call("RPCservice.GetRPC", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Println("Risposta RPC:", reply)
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func Put(lbAddr string) {
	args := Args2{}
	fmt.Print("Insert the Entry Key: ")
	fmt.Scanln(&args.Key)
	fmt.Println(args.Key)

	fmt.Print("Insert the Entry Value: ")
	fmt.Scanln(&args.Value)
	fmt.Println(args.Value)

	var reply *string

	client, _ := HttpConnect(lbAddr)
	err := client.Call("RPCservice.PutRPC", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Println("Risposta RPC:", reply)
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func Update(lbAddr string) {
	args := Args2{}
	fmt.Print("Insert the Key of the Entry to Update: ")
	fmt.Scanln(&args.Key)
	fmt.Println(args.Key)

	fmt.Print("Insert the new Entry Value: ")
	fmt.Scanln(&args.Value)
	fmt.Println(args.Value)

	var reply *string

	client, _ := HttpConnect(lbAddr)
	err := client.Call("RPCservice.UpdateRPC", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Println("Risposta RPC:", reply)
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func Delete(lbAddr string) {
	args := Args1{}
	fmt.Print("Insert the Key of the Entry to Delete: ")
	fmt.Scanln(&args.Key)
	fmt.Println(args.Key)

	var reply *string

	client, _ := HttpConnect(lbAddr)
	err := client.Call("RPCservice.DeleteRPC", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	fmt.Println("Risposta RPC:", reply)
}
