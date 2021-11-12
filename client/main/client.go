package main

// TODO implementare l'operazione di Append
// TODO vedere come in mongo localsys fare la gestione di value values array
// TODO inserire la sovrascrittura nel Put ( che diventa quindi anche update )
// TODO fare la gestione della semantica at-least-once a livello del client
// TODO fare i testing per i due carichi di lavoro visti

import (
	"fmt"
	"io"
	"progetto-sdcc/client/impl"
)

func main() {
Loop:
	for {
		impl.PrintMethodList()
		var cmd string

		fmt.Printf("Inserisci un comando: ")
		_, err := fmt.Scan(&cmd)
		switch {
		case cmd == "1":
			impl.Get()
		case cmd == "2":
			impl.Put()
		case cmd == "3":
			impl.Update()
		case cmd == "4":
			impl.Delete()
		case err == io.EOF:
			break Loop
		}
	}
}
