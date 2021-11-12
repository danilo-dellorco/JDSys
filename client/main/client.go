package main

import (
	"fmt"
	"io"
	"os"
	"progetto-sdcc/client/impl"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: go run client.go LB_IP\n")
	}

	lbAddr := os.Args[1]
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
