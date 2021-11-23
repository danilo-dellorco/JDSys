package impl

import (
	"errors"
	"fmt"
	"net/rpc"
	"progetto-sdcc/utils"
	"strconv"
	"time"
)

var GET string = "Node.GetRPC"
var PUT string = "Node.PutRPC"
var DEL string = "Node.DeleteRPC"
var APP string = "Node.AppendRPC"

/*
Struttura che mantiene i parametri delle RPC
*/
type Args struct {
	Key     string
	Value   string
	Handler string
	Deleted bool
}

/*
Effettua la RPC per la GET
*/
func GetRPC(key string, print bool) {
	args := Args{}
	args.Key = key

	var reply *string

	c := make(chan error)

	client, _ := HttpConnect()
	defer client.Close()
	go CallRPC(GET, client, args, reply, c, print)
	if print {
		rr1_timeout(GET, client, args, reply, c, print)
	}
}

/*
Effettua la RPC per il PUT
*/
func PutRPC(key string, value string, print bool) {
	args := Args{}
	args.Key = key
	args.Value = value

	var reply *string

	c := make(chan error)

	client, _ := HttpConnect()
	defer client.Close()
	go CallRPC(PUT, client, args, reply, c, print)
	rr1_timeout(PUT, client, args, reply, c, print)
}

/*
Effettua la RPC per l'APPEND
*/
func AppendRPC(key string, value string, print bool) {
	args := Args{}
	args.Key = key
	args.Value = value

	var reply *string

	c := make(chan error)

	client, _ := HttpConnect()
	defer client.Close()
	go CallRPC(APP, client, args, reply, c, print)
	rr1_timeout(APP, client, args, reply, c, print)
}

/*
Effettua la RPC per il DELETE
*/
func DeleteRPC(key string, print bool) {
	args := Args{}
	args.Key = key

	var reply *string

	c := make(chan error)

	client, _ := HttpConnect()
	defer client.Close()
	go CallRPC(DEL, client, args, reply, c, print)
	rr1_timeout(DEL, client, args, reply, c, print)
}

/*
Goroutine per l'implementazione della semantica at-least-once.
La ritrasmissione viene effettuata fino a 5 volte, altrimenti si assume che il server sia crashato.
*/
func rr1_timeout(rpc string, client *rpc.Client, args Args, reply *string, c chan error, print bool) {
	signal := make(chan bool)
	res := errors.New("Timeout")
	check := 0
restart_timer:
	for i := 0; i < utils.RR1_RETRIES; i++ {
		go check_timeout(signal)
		select {
		// scade timer per la ritrasmissione
		case <-signal:
			check++
			utils.PrintTs("Timeout elapsed, send new request n°" + strconv.Itoa(check) + "...")
			go CallRPC(rpc, client, args, reply, c, print)

		// arriva risposta dal server
		case res = <-c:
			if res.Error() == "Success" {
				break restart_timer
			}
		}
	}
	//effettuate tutte le ritrasmissioni possibili e non si riceve alcuna risposta
	if check == utils.RR1_RETRIES && res.Error() != "Success" {
		utils.PrintTs("Server unreachable!")
	}
}

/*
Effettua una generica RPC, utilizzata per implementare il meccanismo RR1 per la semantica at-least-once
*/
func CallRPC(rpc string, client *rpc.Client, args Args, reply *string, c chan error, print bool) {
	err := client.Call(rpc, args, &reply)
	defer client.Close()
	if err != nil {
		c <- err
		if print {
			utils.PrintTs("RPC error " + err.Error())
		}
	} else {
		c <- errors.New("Success")
		if print {
			fmt.Println(*reply)
		}
		return
	}
}

func check_timeout(check chan bool) {
	time.Sleep(utils.RR1_TIMEOUT)
	check <- true
}

/*
Permette di instaurare una connessione HTTP con il LB tramite il suo nome DNS.
*/
func HttpConnect() (*rpc.Client, error) {
	i := 0
retry:
	client, err := rpc.DialHTTP("tcp", utils.LB_DNS_NAME+utils.RPC_PORT)
	if err != nil {
		i++
		if i < 10 {
			goto retry
		} else {
			fmt.Println("Connection error: " + err.Error())
		}
	}
	return client, err
}
