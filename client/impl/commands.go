package impl

import (
	"bufio"
	"fmt"
	"os"
	"progetto-sdcc/utils"
	"strings"
)

/*
Permette al client di recuperare il valore associato ad una precisa chiave contattando il LB
*/
func Get() {
	utils.ClearScreen()
	utils.PrintClientTitlebar()
	utils.PrintInBox("GET")
	utils.PrintLineL1()
	key := SecScanln("> Insert the Key of the desired entry")
	utils.PrintLineL1()
	GetRPC(key, true)
	EnterToContinue()
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func Put() {
	utils.ClearScreen()
	utils.PrintClientTitlebar()
	utils.PrintInBox("PUT")
	utils.PrintLineL1()
	key := SecScanln("> Insert the Entry Key")
	value := SecScanln("> Insert the Entry Value")
	utils.PrintLineL1()
	PutRPC(key, value, true)
	EnterToContinue()
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func Append() {
	utils.ClearScreen()
	utils.PrintClientTitlebar()
	utils.PrintInBox("APPEND")
	utils.PrintLineL1()
	key := SecScanln("> Insert the Key of the Entry to Update")
	newValue := SecScanln("> Insert the Value to Append")
	utils.PrintLineL1()
	AppendRPC(key, newValue, true)
	EnterToContinue()
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func Delete() {
	utils.ClearScreen()
	utils.PrintClientTitlebar()
	utils.PrintInBox("DELETE")
	utils.PrintLineL1()
	key := SecScanln("> Insert the Key of the Entry to Delete")
	utils.PrintLineL1()
	DeleteRPC(key, true)
	EnterToContinue()
}

/*
Termina il programma client.
*/
func Exit() {
	utils.PrintLineL1()
	fmt.Println("> Closing Client...")
	fmt.Println("> Goodbye.")
	utils.PrintLineL1()
	fmt.Println("")
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
		} else if arg == "\n" {
		} else {
			break
		}
	}
	return arg[:len(arg)-1]
}

func EnterToContinue() {
	utils.PrintLineL2()
	fmt.Println("")
	fmt.Println("Press the Enter Key to continue...")
	fmt.Scanln()
}
