package impl

import (
	"bufio"
	"fmt"
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
		utils.PrintTs("HTTP Connect error " + err.Error())
		os.Exit(1)
	}
	return client, err
}

/*
Permette al client di recuperare il valore associato ad una precisa chiave contattando il LB
*/
func Get() {
	utils.ClearScreen()
	fmt.Print(utils.PrintInBox("GET"))
	utils.PrintTailerL1()
	key := SecScanln("> Insert the Key of the desired entry")
	utils.PrintTailerL1()
	GetRPC(key)
	EnterToContinue()
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func Put() {
	utils.ClearScreen()
	fmt.Print(utils.PrintInBox("PUT"))
	utils.PrintTailerL1()
	key := SecScanln("> Insert the Entry Key")
	value := SecScanln("> Insert the Entry Value")
	utils.PrintTailerL1()
	PutRPC(key, value)
	EnterToContinue()
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func Append() {
	utils.ClearScreen()
	fmt.Print(utils.PrintInBox("APPEND"))
	utils.PrintTailerL1()
	key := SecScanln("> Insert the Key of the Entry to Update")
	newValue := SecScanln("> Insert the Value to Append")
	utils.PrintTailerL1()
	AppendRPC(key, newValue)
	EnterToContinue()
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func Delete() {
	utils.ClearScreen()
	fmt.Print(utils.PrintInBox("DELETE"))
	utils.PrintTailerL1()
	key := SecScanln("> Insert the Key of the Entry to Delete")
	utils.PrintTailerL1()
	DeleteRPC(key)
	EnterToContinue()
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
		if strings.ContainsAny(arg, "[]{},:./*()\\#") {
			fmt.Println("Inserted value contains not allowed characters []{},:./*()\\#")
			fmt.Println("Retry")
		} else {
			break
		}
	}
	return arg[:len(arg)-1]
}

func EnterToContinue() {
	fmt.Println("Press the Enter Key to continue...")
	fmt.Scanln()
}
