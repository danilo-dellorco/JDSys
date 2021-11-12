package main

import (
	"fmt"
	"progetto-sdcc/utils"
)

func main() {
	utils.GetTimestamp("Prova Timestamp")
}

/*
Esegue un test in cui il workload è composto:
- 85% operazioni di Get
- 15% operazioni di Put
E' possibile specificare tramite il parametro size il numero totali di query da eseguire.
*/
func workload1(size int) {
	numGet := "todo"
	fmt.Println(numGet)
}
