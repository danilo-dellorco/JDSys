package main

// TODO fare i testing per i due carichi di lavoro visti
// TODO verificare il comportamento con la concorrenza. RPC dovrebbe gestirla già da sola, Bisonga vedere Mongo in locale
// come si comporta rispetto ad esempio a due PUT sullo stesso dato.
// TODO migliorare leggibilità del codice
// TODO fare RPC DropDatabase su tutti i nodi per poter fare sempre un test pulito
// TODO testare migrazione e GET da S3
// TODO vedere se si puo dispatchare un nuovo thread ad ogni connessione tcp di file transfer perche cosi non abbiamo concorrenza.
// o comunque vedere se resta in coda e non muore

import (
	"fmt"
	"progetto-sdcc/client/impl"
	"progetto-sdcc/utils"
	"time"
)

func main() {
	for {
		utils.ClearScreen()
		utils.PrintClientTitlebar()
		utils.PrintClientCommandsList()
		var cmd string

		cmd = impl.SecScanln("Insert a Command")
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
			time.Sleep(1 * time.Second)
		}
	}
}
