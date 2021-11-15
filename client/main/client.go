package main

// TODO testare bene la sovrascrittura nel Put ( che diventa quindi anche update)
// TODO fare la gestione della semantica at-least-once a livello del client
// TODO fare i testing per i due carichi di lavoro visti
// TODO verificare il comportamento con la concorrenza. RPC dovrebbe gestirla già da sola, Bisonga vedere Mongo in locale
// come si comporta rispetto ad esempio a due PUT sullo stesso dato.
// TODO Modificare la Delete. Rispetto alle altre operazioni questa non viene aggiornata in automatico con l'update periodico.
// anzi al contrario se vado a cancellare una cosa poi con il merge quella resuscita

import (
	"fmt"
	"io"
	"progetto-sdcc/client/impl"
	"progetto-sdcc/utils"
)

func main() {
	utils.ClearScreen()
Loop:
	for {
		PrintMethodList()
		var cmd string

		fmt.Printf("Insert a Command: ")
		_, err := fmt.Scan(&cmd)
		switch {
		case cmd == "1":
			impl.Get()
		case cmd == "2":
			impl.Put()
		case cmd == "3":
			impl.Delete()
		case cmd == "4":
			impl.Append()
		case cmd == "5":
			impl.Exit()
		default:
			fmt.Println("Command not recognized. Retry.")
		case err == io.EOF:
			break Loop
		}
		utils.ClearScreen()
	}
}

func PrintMethodList() {
	fmt.Println("=============== METHODS LIST ===============")
	fmt.Println("1) Get")
	fmt.Println("2) Put")
	fmt.Println("3) Delete")
	fmt.Println("4) Append")
	fmt.Println("5) Exit")
	fmt.Println("============================================")
}
