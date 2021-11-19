package main

import (
	"fmt"
	"io"
	nodesys "progetto-sdcc/node/impl"
	"progetto-sdcc/utils"
)

func main() {
	utils.ClearScreen()
	node := new(nodesys.Node)
	nodesys.InitNode(node)

	// Ciclo in cui è possibile stampare lo stato attuale del nodo.
Loop:
	for {
		var cmd string
		_, err := fmt.Scan(&cmd)
		switch {

		// Stampa successore e predecessore
		case cmd == "print":
			fmt.Printf("%s", node.ChordClient.String())

		// Stampa la finger table
		case cmd == "fingers":
			fmt.Printf("%s", node.ChordClient.ShowFingers())

		// Stampa la lista di successori
		case cmd == "succ":
			fmt.Printf("%s", node.ChordClient.ShowSucc())

		// Errore
		case err == io.EOF:
			break Loop
		}

	}
	node.ChordClient.Finalize()
	select {}
}
