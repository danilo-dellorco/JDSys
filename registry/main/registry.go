package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"progetto-sdcc/registry/services"
	"progetto-sdcc/utils"
	"time"
)

type term_message struct {
	Status string `json:"status"`
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
	fmt.Println("Healthy Instances:")
	fmt.Println(instances)
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
			// [TODO] Invia un segnale per dirgli che sta terminando e quindi che dovrà
			// inviare il suo DB al successore prima di morire
			sendTerminatingSignal(t.PrivateIP)
		}
		time.Sleep(utils.CHECK_TERMINATING_INTERVAL)
	}
}

func sendTerminatingSignal(ip string) {
	fmt.Println("Sending Terminating Message to node:", ip)
	body := &term_message{Status: "terminating"}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(body)
	url := "http://" + ip + utils.HEARTBEAT_PORT
	req, _ := http.NewRequest("POST", url, buf)
	req.Close = true

	client := &http.Client{}
	res, e := client.Do(req)
	fmt.Println(res)
	fmt.Println(res.StatusCode)
	if e != nil {
		log.Fatal(e)
	}
	//defer res.Body.Close()

	_, e = ioutil.ReadAll(res.Body)
	if e != nil {
		log.Fatal(e)
	} else {
		// Print the body to the stdout
		io.Copy(os.Stdout, res.Body)
	}
}

func sendTest(ip string) *http.Response {
	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	url := "http://" + ip + utils.HEARTBEAT_PORT
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return resp
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Wrong usage: Specify user \"d\" or \"j\"")
		return
	}
	services.SetupUser()
	go checkTerminatingNodes()
	fmt.Printf("Server Waiting For Connection... \n")
	service := InitializeService()
	rpc.Register(service)
	rpc.HandleHTTP()
	log.Fatal(http.ListenAndServe(":1234", nil))
}
