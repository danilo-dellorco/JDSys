package main

import (
	"progetto-sdcc/utils"
	"strconv"
)

var LB_ADDR = "Address Of Network LB"

var PERC_75 float32 = 0.75
var PERC_15 float32 = 0.15

var PERC_40 float32 = 0.40
var PERC_20 float32 = 0.20

func main() {
	utils.GetTimestamp()
}

/*
Esegue un test in cui il workload è composto:
- 85% operazioni di Get
- 15% operazioni di Put
E' possibile specificare tramite il parametro size il numero totali di query da eseguire.
*/
func workload1(size float32) {
	numGet := int(PERC_75 * size)
	numPut := int(PERC_15 * size)

	go runGetQueries(numGet)
	go runPutQueries(numPut)
}

func runGetQueries(num int) {
	for i := 0; i < num; i++ {
		key := "test_key_" + strconv.Itoa(i)
		go TestGet(key)
	}
}

func runPutQueries(num int) {
	for i := 0; i < num; i++ {
		key := "test_key_" + strconv.Itoa(i)
		value := "test_value_" + strconv.Itoa(i)
		go TestPut(key, value)
	}
}
