package impl

import (
	"fmt"
	"log"
	"net/rpc"
	"progetto-sdcc/utils"
	"strings"
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
	key := SecScanln("Insert the Desired Key")

	GetRPC(key)
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func Put() {

	key := SecScanln("Insert the Entry Key")

	value := SecScanln("Insert the Entry Value")

	PutRPC(key, value)
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func Append() {
	key := SecScanln("Insert the Key of the Entry to Update")

	newValue := SecScanln("Insert the Value to Append")

	AppendRPC(key, newValue)
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func Delete() {
	key := SecScanln("Insert the Key of the Entry to Delete")

	DeleteRPC(key)
}

/*
Prende input da tastiera in modo sicuro, rimuovendo eventuali caratteri che potrebbero
permettere ad un attaccante di effettuare una Injection
*/
func SecScanln(message string) string {
	var arg string

	for {
		fmt.Print(message + ": ")
		fmt.Scanln(&arg)
		if strings.ContainsAny(arg, "[]{},:./*") {
			fmt.Println("Inserted value contains not allowed characters")
			fmt.Println("Retry")
		} else {
			break
		}
	}
	return arg
}
