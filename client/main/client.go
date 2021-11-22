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
	"io"
	"progetto-sdcc/client/impl"
	"progetto-sdcc/utils"
	"strings"
	"time"
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
		case err == io.EOF:
			break Loop
		default:
			fmt.Println("Command not recognized. Retry.")
		}

		// Aspettiamo per far stampare prima la risposta se arriva in ritardo, per poi pulire lo schermo
		time.Sleep(3 * time.Second)
		utils.ClearScreen()
	}
}

func PrintMethodList() {
	utils.PrintHeaderL1("SDCC Distributed Key-Value Storage")
	utils.PrintTailerL1()
	fmt.Print(utils.StringInBox("COMMANDS LIST"))

	get := "Get"
	put := "Put"
	del := "Delete"
	app := "Append"
	ext := "Exit"

	top := "+" + strings.Repeat("—", 15) + "+\n"
	row1 := "| 1 |  " + get + strings.Repeat(" ", 3) + "   |\n"
	row2 := "| 2 |  " + put + strings.Repeat(" ", 3) + "   |\n"
	row3 := "| 3 |  " + del + strings.Repeat(" ", 0) + "   |\n"
	row4 := "| 4 |  " + app + strings.Repeat(" ", 0) + "   |\n"
	row5 := "| 5 |  " + ext + strings.Repeat(" ", 2) + "   |\n"
	bottom := top

	fmt.Println(top + row1 + row2 + row3 + row4 + row5 + bottom)
}
