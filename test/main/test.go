package main

import (
	"fmt"
	"os"
	"progetto-sdcc/test/impl"
	"progetto-sdcc/utils"
	"strconv"
	"time"
)

var PERC_75 float32 = 0.75
var PERC_15 float32 = 0.15
var PERC_40 float32 = 0.40
var PERC_20 float32 = 0.20

var WORKLOAD []int

func main() {
	if len(os.Args) != 3 {
		fmt.Println("You need to specify the workload type to test.")
		fmt.Println("Usage: go run test.go WORKLOAD SIZE")
		return
	}
	fmt.Println("Test PID:", os.Getpid())
	test_type := os.Args[1]
	test_size_int, _ := strconv.Atoi(os.Args[2])
	test_size := float32(test_size_int)

	switch test_type {
	case "workload1":
		workload1(test_size)
		time.Sleep(utils.TEST_STEADY_TIME)
		utils.PrintHeaderL3("System it's at steady-state")
		//measureResponseTime()
	case "workload2":
		workload2(test_size)
	}
	select {}
}

/*
Effettua una richiesta di Put, una di Update, una di Get, una di Append e una di Delete, misurando poi il tempo medio di risposta
*/
func measureResponseTime() {
	utils.PrintHeaderL2("Starting Measuring Response Time")
	rt1 := impl.TestPut("rt_key", "rt_value", true, 0)
	rt2 := impl.TestPut("rt_key", "rt_value_upd", true, 0)
	rt3 := impl.TestGet("rt_key", true, 0)
	rt4 := impl.TestAppend("rt_key", "rt_value_app", true, 0)
	rt5 := impl.TestDelete("rt_key", true)

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

	impl.WORKLOAD_GET = make([]int, numGet)
	impl.WORKLOAD_PUT = make([]int, numPut)

	utils.PrintHeaderL2("Start Spawning Threads for Workload 1")
	utils.PrintStringInBoxL2("# Get | "+strconv.Itoa(numGet), "# Put | "+strconv.Itoa(numPut))
	utils.PrintLineL2()
	go runPutQueries(numPut)
	go runGetQueries(numGet)
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

	impl.WORKLOAD_GET = make([]int, numGet)
	impl.WORKLOAD_PUT = make([]int, numPut)
	impl.WORKLOAD_APP = make([]int, numApp)

	utils.PrintHeaderL2("Start Spawning Threads for Workload 2")
	utils.PrintStringInBoxL2("# Get | "+strconv.Itoa(numGet), "# Put | "+strconv.Itoa(numPut))
	utils.PrintLineL2()

	go runGetQueries(numGet)
	go runPutQueries(numPut)
	go runAppendQueries(numApp)
}

func runGetQueries(num int) {
	id := 0
	for {
		if id == num {
			id = 0
		}
		key := "test_key_" + strconv.Itoa(id)
		if impl.WORKLOAD_GET[id] != 1 {
			go impl.TestGet(key, false, id)
		}
		id++
	}
}

func runPutQueries(num int) {
	id := 0
	for {
		if id == num {
			id = 0
		}
		key := "test_key_" + strconv.Itoa(id)
		value := "test_value_" + strconv.Itoa(id)
		if impl.WORKLOAD_GET[id] != 1 {
			go impl.TestPut(key, value, false, id)
		}
		id++
	}
}

func runAppendQueries(num int) {
	id := 0
	for {
		key := "test_key_" + strconv.Itoa(id)
		arg := "_test_arg_" + strconv.Itoa(id)
		go impl.TestAppend(key, arg, false, id)
	}
}
