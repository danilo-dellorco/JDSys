package main

import (
	"fmt"
	"io"
	nodesys "progetto-sdcc/node/impl"
)

func main() {
	node := new(nodesys.Node)
	nodesys.InitNode(node)

Loop:
	for {
		var cmd string
		_, err := fmt.Scan(&cmd)
		switch {
		case cmd == "print":
			//stampa successore e predecessore
			fmt.Printf("%s", node.ChordClient.String())
		case cmd == "fingers":
			//stampa la finger table
			fmt.Printf("%s", node.ChordClient.ShowFingers())
		case cmd == "succ":
			//stampa la lista di successori
			fmt.Printf("%s", node.ChordClient.ShowSucc())
		case err == io.EOF:
			break Loop
		}

	}
	node.ChordClient.Finalize()
	select {}
}
