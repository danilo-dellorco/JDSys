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
Permette di instaurare una connessione HTTP con il LB tramite il suo nome DNS.
*/
func HttpConnect() (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", utils.LB_DNS_NAME+utils.RPC_PORT)
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
func Get() {
	var key string
	fmt.Print("Insert the Desired Key: ")
	fmt.Scanln(&key)

	GetRPC(key)
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func Put() {
	var key string
	var value string
	fmt.Print("Insert the Entry Key: ")
	fmt.Scanln(&key)

	fmt.Print("Insert the Entry Value: ")
	fmt.Scanln(&value)

	PutRPC(key, value)
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func Update() {
	var key string
	var newValue string
	fmt.Print("Insert the Key of the Entry to Update: ")
	fmt.Scanln(&key)

	fmt.Print("Insert the new Entry Value: ")
	fmt.Scanln(&newValue)

	UpdateRPC(key, newValue)
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func Delete() {
	var key string
	fmt.Print("Insert the Key of the Entry to Delete: ")
	fmt.Scanln(&key)

	DeleteRPC(key)
}
