package main

import (
	"progetto-sdcc/client/impl"
	"progetto-sdcc/utils"
	"time"
)

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
func TestGet(key string) time.Duration {
	start := utils.GetTimestamp()

	impl.GetRPC(key)

	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func TestPut(key string, value string) time.Duration {
	start := utils.GetTimestamp()

	impl.PutRPC(key, value)

	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func TestAppend(key string, value string) time.Duration {
	start := utils.GetTimestamp()

	impl.AppendRPC(key, value)

	end := utils.GetTimestamp()

	return end.Sub(start)
}

/*
Permette al client di eliminare una coppia key-value dal sistema di storage contattando il LB
*/
func TestDelete(key string) time.Duration {
	start := utils.GetTimestamp()

	impl.DeleteRPC(key)

	end := utils.GetTimestamp()

	return end.Sub(start)
}
