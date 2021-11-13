package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"progetto-sdcc/registry/services"
	"progetto-sdcc/utils"
	"time"
)

func main() {
	go checkTerminatingNodes()
	fmt.Printf("Server Registry Waiting For Incoming Connection... \n")
	service := InitializeService()
	rpc.Register(service)
	rpc.HandleHTTP()
	log.Fatal(http.ListenAndServe(":1234", nil))
}

/*
Struttura per il passaggio dei parametri alla RPC
*/
type Args struct{}

/*
Pseudo-Interfaccia che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici
*/
type DHThandler int

/*
Un nodo, per effettuare Create/Join, deve conoscere i nodi presenti nell'anello
*/
func (s *DHThandler) JoinRing(args *Args, reply *[]string) error {
	instances := checkActiveNodes()
	var list = make([]string, len(instances))
	for i := 0; i < len(instances); i++ {
		list[i] = instances[i].PrivateIP
	}
	*reply = list
	return nil
}

/*
Inizializza il servizio DHT
*/
func InitializeService() *DHThandler {
	service := new(DHThandler)
	return service
}

/*
Restituisce tutte le istanze healthy presenti
*/
func checkActiveNodes() []services.Instance {
	instances := services.GetActiveNodes()
	//fmt.Println("Healthy Instances:")
	//fmt.Println(instances)
	return instances
}

/*
Controlla ogni tot secondi quali sono le istanze in terminaione. Invia a queste un segnale in modo che prima
di terminare possano inviare le proprie entry ad un altro nodo
*/
func checkTerminatingNodes() {
	fmt.Println("Starting Check Terminating Nodes Routine....")
	go services.Start_cache_flush_service()
	for {
		terminating := services.GetTerminatingInstances()
		for _, t := range terminating {
			sendTerminatingSignalRPC(t.PrivateIP)
		}
		time.Sleep(utils.CHECK_TERMINATING_INTERVAL)
	}
}

/*
Chiamata a RPC che invia il segnale di terminazione ad un nodo schedulato per la terminazione
*/
func sendTerminatingSignalRPC(ip string) {
	fmt.Println("Sending Terminating Message to node:", ip)
	client, err := rpc.DialHTTP("tcp", ip+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var reply string
	args := Args{}
	err = client.Call("RPCservice.TerminateInstanceRPC", args, &reply)
	if err != nil {
		log.Fatal("GetRPC error:", err)
	}
	defer client.Close()
	fmt.Println("Risposta RPC:", reply)
}
