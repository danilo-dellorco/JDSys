package main

import (
	"context"
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

/*
Struttura per il passaggio dei parametri alla RPC
*/
type Args struct {
	Handler string
	Deleted bool
}

/*
Servizio per le RPC del registry
*/
type DHThandler int

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	utils.ClearScreen()
	server := InitRegistry()

	//Aspetta segnali per chiudere tutte le connessioni al Ctrl+C
	<-done
	utils.PrintTs("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		utils.PrintTs("Server Shutdown Failed: " + err.Error())
		os.Exit(1)
	}
	utils.PrintTs("Server Exited Properly")
}

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
func InitializeDHTService() *DHThandler {
	service := new(DHThandler)
	return service
}

/*
Restituisce tutte le istanze healthy presenti
*/
func checkActiveNodes() []amazon.Instance {
	instances := amazon.GetActiveNodes()
	return instances
}

/*
Controlla periodicamente quali sono le istanze in terminazione. Invia a queste un segnale in modo che prima
di terminare possano inviare le proprie entry ad un altro nodo
*/
func checkTerminatingNodes() {
	utils.PrintHeaderL2("Starting Checking Terminating Nodes")
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
	utils.PrintTs("Sending Terminating Message to node: " + ip)
	client, err := rpc.DialHTTP("tcp", ip+utils.RPC_PORT)
	if err != nil {
		utils.PrintTs("dialing: " + err.Error())
		os.Exit(1)
	}
	var reply string
	args := Args{}
	err = client.Call("Node.TerminateInstanceRPC", args, &reply)
	if err != nil {
		utils.PrintTs("TerminateInstanceRPC error: " + err.Error())
		os.Exit(1)
	}
	utils.PrintTs(ip + ": " + reply)
}

/*
Avvia periodicamente il processo iterativo di scambio di aggiornamenti tra un nodo e il suo successore per la riconciliazione.
Il processo permette di raggiungere la consistenza finale se non si verificano aggiornamenti in questa finestra temporale
*/
// TODO calcolare bene il valore della finestra temporale per la relazione
func startPeriodicUpdates() {
	utils.PrintHeaderL2("Starting periodic updates for reconciliation Routine")
	for {
		time.Sleep(utils.START_CONSISTENCY_INTERVAL)
	retry:
		nodes := checkActiveNodes()
		if len(nodes) == 0 || len(nodes) == 1 {
			utils.PrintTs("Wait the correct construction of the DHT to start the updates routine of the ring")
			time.Sleep(utils.WAIT_SUCC_TIME)
			goto retry
		}
		// Recuperate tutte le istanze attive, si invia la richiesta ad un nodo a caso
		var list = make([]string, len(nodes))
		for i := 0; i < len(nodes); i++ {
			list[i] = nodes[i].PrivateIP
		}
		utils.PrintTs("Choosing random node to start the reconciliation")
		startReconciliationRPC(list[rand.Intn(len(list))])
	}
}

/*
Invocazione dell'RPC che avvia lo scambio di aggiornamenti tra i nodi per raggiungere la consistenza finale
*/
func startReconciliationRPC(ip string) {
	utils.PrintTs("Sending db exchange signal to node: " + ip)
	client, err := rpc.DialHTTP("tcp", ip+utils.RPC_PORT)
	if err != nil {
		utils.PrintTs("dialing: " + err.Error())
	}
	var reply string
	args := Args{}
	args.Handler = ""
	args.Deleted = false
	err = client.Call("Node.ConsistencyHandlerRPC", args, &reply)
	if err != nil {
		utils.PrintTs("ConsistencyHandlerRPC error: " + err.Error())
	}
}

func InitRegistry() *http.Server {
	utils.PrintHeaderL1("REGISTRY SETUP")

	server := &http.Server{
		Addr:    utils.REGISTRY_PORT,
		Handler: http.DefaultServeMux,
	}
	service := InitializeDHTService()
	rpc.Register(service)
	rpc.HandleHTTP()

	go server.ListenAndServe()
	utils.PrintTs("Service Registry waiting for incoming connections")

	go checkTerminatingNodes()
	go startPeriodicUpdates()
	return server
}
