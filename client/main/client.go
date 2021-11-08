package main

import (
	"fmt"
	"os"
	"progetto-sdcc/client/impl"
)

func main() {
	var elbAddress string
	elbAddress = os.Args[1]
	fmt.Println(elbAddress)
	if len(os.Args) != 2 {
		fmt.Printf("Usage: go run client.go SERVER_IP\n")
	}
	for {
		impl.PrintMethodList()

		var cmd string

		fmt.Printf("Inserisci un comando: ")
		fmt.Scanln(&cmd)

		switch cmd {
		case "1":
			impl.Get()
		}
	}
}
