package impl

import (
	"fmt"
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
func TestGet(key string, channel chan bool, print bool) time.Duration {
	// per la Get effettiva tramite cui misuriamo il tempo di risposta non abbiamo bisogno del thread che mantiene stabile il carico
	if channel != nil {
		go CheckGet(key, channel, print)
	}

	start := utils.GetTimestamp()
	impl.GetRPC(key, print)
	end := utils.GetTimestamp()

	channel <- true

	return end.Sub(start)
}

/*
Permette al client di inserire una coppia key-value nel sistema di storage contattando il LB
*/
func TestPut(key string, value string, channel chan bool, print bool) time.Duration {
	// per la Put effettiva tramite cui misuriamo il tempo di risposta non abbiamo bisogno del thread che mantiene stabile il carico
	fmt.Println("nuovo thread")
	if channel != nil {
		go CheckPut(key, value, channel, print)
	}

	start := utils.GetTimestamp()
	impl.PutRPC(key, value, print)
	end := utils.GetTimestamp()

	channel <- true

	return end.Sub(start)
}

/*
Permette al client di aggiornare una coppia key-value presente nel sistema di storage contattando il LB
*/
func TestAppend(key string, value string, print bool) time.Duration {
	start := utils.GetTimestamp()

	impl.AppendRPC(key, value, print)

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

func CheckGet(key string, channel chan bool, print bool) {
	for {
		end := <-channel
		if end {
			go TestGet(key, channel, print)
		}
	}
}

func CheckPut(key string, value string, channel chan bool, print bool) {
	for {
		end := <-channel
		if end {
			fmt.Println("spawn nuovo thread")
			go TestPut(key, value, channel, print)
		}
	}
}
