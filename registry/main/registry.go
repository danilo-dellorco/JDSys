package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"progetto-sdcc/registry/services"
	"time"
)

var instances []services.Instance

// Struttura per il passaggio dei parametri alla RPC
type Args struct{}

// Pseudo-Interfaccia che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
// ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici!
type DHThandler int

// Metodo 1 dell'interfaccia
func (s *DHThandler) JoinRing(args *Args, reply *[]string) error {
	var list = make([]string, 10)
	for i := 0; i < len(instances); i++ {
		list[i] = instances[i].PrivateIP
	}
	*reply = list
	return nil
}

func InitializeService() *DHThandler {
	service := new(DHThandler)
	return service
}

func checkActiveNodes() {
	for {
		instances = services.GetActiveNodes()
		fmt.Println("Info Healthy Instances:")
		fmt.Println(instances)
		time.Sleep(time.Second * 10)
	}
}

func checkTerminatingNodes() {
	for {
		terminating := services.GetTerminatingInstances()
		for t := range terminating {
			// [TODO] Invia un segnale per dirgli che sta terminando e quindi che dovrà
			// inviare il suo DB al successore prima di morire
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Wrong usage: Specify user \"d\" or \"j\"")
		return
	}
	services.SetupUser()
	go checkActiveNodes()
	//diocane := services.GetActiveNodes()
	//fmt.Println("Info Healthy Instances:")
	//fmt.Println(diocane)
	fmt.Printf("Server Waiting For Connection... \n")
	service := InitializeService()
	rpc.Register(service)
	rpc.HandleHTTP()
	log.Fatal(http.ListenAndServe(":1234", nil))
}
