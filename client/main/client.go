package main

import (
	"fmt"
	"os"
	"progetto-sdcc/client/impl"
)

func main() {
	var serverAddress string
	serverAddress = os.Args[1]
	if len(os.Args) != 2 {
		fmt.Printf("Usage: go run client.go SERVER_IP\n")
	}
	for {
		var cmd string
		fmt.Printf("Inserisci un comando: ")
		fmt.Scanln(&cmd)

		switch cmd {
		case "list":
			impl.GetMethodsList(serverAddress)
		}
	}
}
