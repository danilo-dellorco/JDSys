package main

import (
	"fmt"
	"os"
	mongo "progetto-sdcc/node/mongo/api"
	"progetto-sdcc/utils"
	"strconv"
	"time"
)

var PERC_75 float32 = 0.75
var PERC_15 float32 = 0.15
var PERC_40 float32 = 0.40
var PERC_20 float32 = 0.20

func main() {
	if len(os.Args) != 2 {
		fmt.Println("You need to specify the workload type to test.")
		fmt.Println("Usage: go run test.go WORKLOAD SIZE")
		return
	}
	test_type := os.Args[1]
	test_size_int, _ := strconv.Atoi(os.Args[2])
	test_size := float32(test_size_int)

	switch test_type {
	case "workload1":
		workload1(test_size)
		time.Sleep(utils.TEST_STEADY_TIME)
		measureResponseTime()
	case "workload2":
		workload2(test_size)
	}
	select {}
}

func localPutTest(mongo mongo.MongoInstance) {
	i := 0
	for {
		go mongo.PutEntry("key_test", strconv.Itoa(i))
		i++
	}
}

func localGetTest(mongo mongo.MongoInstance) {
	i := 0
	for {
		go mongo.GetEntry("key_test")
		i++
	}
}

/*
Effettua una richiesta di Put, una di Update, una di Get, una di Append e una di Delete, misurando poi il tempo medio di risposta
*/
func measureResponseTime() {
	rt1 := TestPut("rt_key", "rt_value")
	rt2 := TestPut("rt_key", "rt_value_upd")
	rt3 := TestGet("rt_key")
	rt4 := TestAppend("rt_key", "rt_value_app")
	rt5 := TestDelete("rt_key")

	total := rt1 + rt2 + rt3 + rt4 + rt5
	meanRt := total / 5
	fmt.Println("Mean Response Time:", meanRt)
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
		key := "test_key_" + strconv.Itoa(i)
		arg := "_test_arg_" + strconv.Itoa(i)
		go TestAppend(key, arg)
	}
}
