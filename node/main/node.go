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
	utils.PrintHeaderL1("NODE SYSTEM")

	// Ciclo in cui è possibile stampare lo stato attuale del nodo.
Loop:
	for {
		var cmd string
		_, err := fmt.Scan(&cmd)
		var s string
		switch {
		// Stampa successore e predecessore
		case cmd == "print":
			s = fmt.Sprintf("%s", node.ChordClient.String())
			utils.PrintTs(s)
		// Stampa la finger table
		case cmd == "fingers":
			s = fmt.Sprintf("%s", node.ChordClient.ShowFingers())
			utils.PrintTs(s)
		// Stampa la lista di successori
		case cmd == "succ":
			s = fmt.Sprintf("%s", node.ChordClient.ShowSucc())
			utils.PrintTs(s)
		// Errore
		case err == io.EOF:
			break Loop
		}

	}
	node.ChordClient.Finalize()
	select {}
}
