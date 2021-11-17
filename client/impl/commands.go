package impl

import (
	"bufio"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"progetto-sdcc/utils"
	"strings"
)

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
	fmt.Println("va: ", value)
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
Termina il programma client.
*/
func Exit() {
	fmt.Println("Closing Client...")
	fmt.Println("Goodbye.")
	os.Exit(0)
}

/*
Prende input da tastiera in modo sicuro, rimuovendo eventuali caratteri che potrebbero
permettere ad un attaccante di effettuare una Injection su MongoDB
*/
func SecScanln(message string) string {
	arg := ""
	for {
		fmt.Print(message + ": ")
		arg, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.ContainsAny(arg, "[]{},:./*") {
			fmt.Println("Inserted value contains not allowed characters")
			fmt.Println("Retry")
		} else {
			break
		}
	}
	return arg[:len(arg)-1]
}
