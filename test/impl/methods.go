package impl

import (
	"progetto-sdcc/client/impl"
	"progetto-sdcc/utils"
	"time"
)

var WORKLOAD_GET []int
var WORKLOAD_PUT []int
var WORKLOAD_APP []int

/*
Parametri per le operazioni di Get e Delete
*/
type Args1 struct {
	Key string
}

/*
Parametri per le operazioni di Put e Update
*/
type Args2 struct {
	Key   string
	Value string
}

/*
Permette al client di recuperare il valore associato ad una precisa chiave contattando il LB
*/
func TestGet(key string, print bool, id int) time.Duration {
	WORKLOAD_GET[id] = 1

	start := utils.GetTimestamp()
	impl.GetRPC(key, print)
	WORKLOAD_GET[id] = 0
	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func TestPut(key string, value string, print bool, id int) time.Duration {
	WORKLOAD_PUT[id] = 1
	start := utils.GetTimestamp()

	impl.PutRPC(key, value, print)

	WORKLOAD_PUT[id] = 0

	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func TestAppend(key string, value string, print bool, id int) time.Duration {
	WORKLOAD_APP[id] = 1

	start := utils.GetTimestamp()

	impl.AppendRPC(key, value, print)

	WORKLOAD_APP[id] = 0

	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func TestDelete(key string, print bool) time.Duration {
	start := utils.GetTimestamp()

	impl.DeleteRPC(key, print)

	end := utils.GetTimestamp()

	return end.Sub(start)
}
