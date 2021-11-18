package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"progetto-sdcc/registry/amazon"
	"progetto-sdcc/utils"
	"syscall"
	"time"
)

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:    utils.REGISTRY_PORT,
		Handler: http.DefaultServeMux,
	}

	go checkTerminatingNodes()

	fmt.Printf("Server Registry Waiting For Incoming Connection... \n")
	service := InitializeService()
	rpc.Register(service)
	rpc.HandleHTTP()
	go server.ListenAndServe()

	go startPeriodicUpdates()

	//Aspetta segnali per chiudere tutte le connessioni al Ctrl+C
	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}

/*
Struttura per il passaggio dei parametri alla RPC
*/
type Args struct {
	Handler string
	Deleted bool
}

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
func checkActiveNodes() []amazon.Instance {
	instances := amazon.GetActiveNodes()
	//fmt.Println("Healthy Instances:")
	//fmt.Println(instances)
	return instances
}

/*
Controlla periodicamente quali sono le istanze in terminazione. Invia a queste un segnale in modo che prima
di terminare possano inviare le proprie entry ad un altro nodo
*/
func checkTerminatingNodes() {
	fmt.Println("Starting Check Terminating Nodes Routine....")
	go amazon.Start_cache_flush_service()
	for {
		terminating := amazon.GetTerminatingInstances()
		for _, t := range terminating {
			sendTerminatingSignalRPC(t.PrivateIP)
		}
		time.Sleep(utils.CHECK_TERMINATING_INTERVAL)
	}
}

/*
Invocazione dell'RPC che invia il segnale di terminazione ad un nodo schedulato per la terminazione
*/
func sendTerminatingSignalRPC(ip string) {
	fmt.Println("Sending Terminating Message to node:", ip)
	client, err := rpc.DialHTTP("tcp", ip+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var reply string
	args := Args{}
	err = client.Call("Node.TerminateInstanceRPC", args, &reply)
	if err != nil {
		log.Fatal("GetRPC error:", err)
	}
	//defer client.Close()
	fmt.Println("Risposta RPC:", reply)
}

/*
Avvia periodicamento il processo iterativo di scambio di aggiornamenti tra un nodo e il suo successore per la riconciliazione.
Il processo permette di raggiungere la consistenza finale se non si verificano aggiornamenti in questa finestra temporale
*/
func startPeriodicUpdates() {
	time.Sleep(time.Minute)
	fmt.Println("Starting periodic updates for reconciliation Routine....")
retry:
	for {
		nodes := checkActiveNodes()
		if len(nodes) == 0 || len(nodes) == 1 {
			fmt.Println("Wait the correct construction of the DHT to start the updates routine of the ring")
			time.Sleep(10 * time.Second)
			goto retry
		}
		//recuperate tutte le istanze attive, si invia la richiesta ad un nodo a caso
		var list = make([]string, len(nodes))
		for i := 0; i < len(nodes); i++ {
			list[i] = nodes[i].PrivateIP
		}
		startFinalConsistencyRPC(list[rand.Intn(len(list))])
		time.Sleep(utils.START_CONSISTENCY_INTERVAL)
	}
}

/*
Invocazione dell'RPC che avvia lo scambio di aggiornamenti tra i nodi per raggiungere la consistenza finale
*/
func startFinalConsistencyRPC(ip string) {
	fmt.Println("Sending signal to start DB exchange to node:", ip)
	client, err := rpc.DialHTTP("tcp", ip+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	var reply string
	args := Args{}
	args.Handler = ""
	args.Deleted = false
	err = client.Call("Node.ConsistencyHandlerRPC", args, &reply)
	if err != nil {
		log.Fatal("GetRPC error:", err)
	}
	//defer client.Close()
	fmt.Println("Risposta RPC:", reply)
}
