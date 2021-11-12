package main

import (
	"fmt"
	"io"
	"os"
	"progetto-sdcc/client/impl"
	"progetto-sdcc/utils"
)

// Mantiene l'indirizzo DNS del Load Balancer
var lbAddr string

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Wrong usage: Specify LB DNS name with \"d\" or \"j\"\n")
	}
	// TODO rimuovere le differenze danilo / jacopo
	user := os.Args[1]
	if user == "d" {
		lbAddr = utils.LB_DNS_NAME_D
	} else {
		lbAddr = utils.LB_DNS_NAME_J
	}

Loop:
	for {
		impl.PrintMethodList()
		var cmd string

		fmt.Printf("Inserisci un comando: ")
		_, err := fmt.Scan(&cmd)
		switch {
		case cmd == "1":
			impl.Get(lbAddr)
		case cmd == "2":
			impl.Put(lbAddr)
		case cmd == "3":
			impl.Update(lbAddr)
		case cmd == "4":
			impl.Delete(lbAddr)
		case err == io.EOF:
			break Loop
		}
	}
}
