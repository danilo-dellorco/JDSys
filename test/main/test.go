package main

import (
	"fmt"
	"os"
	mongo "progetto-sdcc/node/localsys"
	"progetto-sdcc/node/localsys/structures"
	"strconv"
)

var LB_ADDR = "Address Of Network LB"

var PERC_75 float32 = 0.75
var PERC_15 float32 = 0.15

var PERC_40 float32 = 0.40
var PERC_20 float32 = 0.20

func main() {
	if len(os.Args) != 2 {
		fmt.Println("You need to specify the workload type to test.")
		fmt.Println("Usage: go run test.go WORKLOAD")
		return
	}
	test_type := os.Args[1]

	mc := mongo.InitLocalSystem()
	switch test_type {
	case "put":
		localPutTest(mc)
	case "get":
		localGetTest(mc)

	}
	select {}
}

func localPutTest(mongo structures.MongoClient) {
	i := 0
	for {
		go mongo.PutEntry("key_test", strconv.Itoa(i))
		i++
	}
}

func localGetTest(mongo structures.MongoClient) {
	i := 0
	for {
		go mongo.GetEntry("key_test")
		i++
	}
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

/*
Esegue un test in cui il workload è composto:
- 40% operazioni di Get
- 40% operazioni di Put
- 20% operazioni di Append
E' possibile specificare tramite il parametro size il numero totali di query da eseguire.
*/
func workload2(size float32) {
	numGet := int(PERC_40 * size)
	numPut := int(PERC_40 * size)
	numApp := int(PERC_20 * size)

	go runGetQueries(numGet)
	go runPutQueries(numPut)
	go runAppendQueries(numApp)
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

func runAppendQueries(num int) {
	for i := 0; i < num; i++ {
		//key := "test_key_" + strconv.Itoa(i)
		//arg := "_test_arg_" + strconv.Itoa(i)
		//go TestAppend(key, value)
	}
}
