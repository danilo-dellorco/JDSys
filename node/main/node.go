package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	chord "progetto-sdcc/node/chord/net"
	mongo "progetto-sdcc/node/localsys"
	"progetto-sdcc/node/localsys/structures"
	"progetto-sdcc/utils"
	"time"
)

type EmptyArgs struct{}

var mongoClient structures.MongoClient
var me *chord.ChordNode

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Wrong usage: Specify registry private IP address")
		return
	}

	// TODO invece che aspettare 40 secondi forse dopo aver farto partire il listener degli heartbeat
	// possiamo inizializzare il database locale invece di fare una sleep facciamo tutta la config locale che comunque
	// ci mette tempo!!
	InitHealthyNode()
	InitChordDHT()
	//service := InitServiceRPC()

	//rpc.Register(service)
	//rpc.HandleHTTP()
	//service.ListenHttpConnection()

	// [TODO] Togliere, sono stampe di debug ma il nodo non riceve comandi da riga di comando ma tramite RPC
Loop:
	for {
		var cmd string
		_, err := fmt.Scan(&cmd)
		switch {
		case cmd == "print":
			//print out successor and predecessor
			fmt.Printf("%s", me.String())
		case cmd == "fingers":
			//print out finger table
			fmt.Printf("%s", me.ShowFingers())
		case cmd == "succ":
			//print out successor list
			fmt.Printf("%s", me.ShowSucc())
		case err == io.EOF:
			break Loop
		}

	}
	me.Finalize()
	select {}
}

type term_message struct {
	Status string
}

/*
Gestisce gli hearthbeat del Load Balancer ed i messaggi di Terminazione dal Service Registry
*/
func terminate_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")
	if r.Method == "POST" {
		fmt.Println("Ricevuta Richiesta di Post!")
		var m term_message
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			panic(err)
		}
		status := m.Status
		if status == "terminating" {

			// Invio al nodo successore l'intero database del nodo in terminazione
			fmt.Println("Node Scheduled to Terminating...")
			succ := me.GetSuccessor()
			ip := succ.GetIpAddr()
			mongo.SendUpdate(mongoClient, ip)
		}
	}
}

/*
Inizializza un listener sulla porta 8888, su cui il Nodo riceve gli HeartBeat del Load Balancer,
ed i segnali di terminazione dal service registry.
*/
func StartHeartBeatListener() {
	fmt.Println("Start Listening Messages on port:", utils.HEARTBEAT_PORT)
	http.HandleFunc("/", terminate_handler)
	http.ListenAndServe(utils.HEARTBEAT_PORT, nil)
}

/*
Restituisce l'indirizzo IP in uscita preferito della macchina che hosta il nodo
*/
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

/*
Permette di instaurare una connessione HTTP con il server all'indirizzo specificato.
*/
func HttpConnect(serverAddress string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", serverAddress+":1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return client, err
}

/*
Permette al nodo di inserirsi nell'anello chord contattando il server specificato
*/
func JoinDHT(serverAddress string) []string {
	args := EmptyArgs{}
	var reply []string

	client, _ := HttpConnect(serverAddress)
	err := client.Call("DHThandler.JoinRing", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	return reply
}

/*
Permette al nodo di essere rilevato come Healthy Instance dal Load Balancer.
Inizia anche una routine che è sempre in ascolto per la ricezione degli HeartBeat
*/
func InitHealthyNode() {
	// Inizia a ricevere gli HeartBeat
	go StartHeartBeatListener()

	// Inizia a configurare il sistema di storage locale
	// TODO decommentare
	mongoClient = mongo.InitLocalSystem()

	// Attende di diventare healthy per il Load Balancer
	fmt.Println("Waiting for ELB Health Checking...")
	time.Sleep(utils.NODE_HEALTHY_TIME)
	fmt.Println("EC2 Node Up & Running")
}

func InitChordDHT() {
	fmt.Println("Initializing Chord DHT")
	// Setup dei Flags
	addressPtr := flag.String("addr", "", "the port you will listen on for incomming messages")
	joinPtr := flag.String("join", "", "an address of a server in the Chord network to join to")
	flag.Parse()

	// Ottiene l'indirizzo IP dell'host utilizzato nel VPC
	*addressPtr = GetOutboundIP().String() + ":4567"
	me = new(chord.ChordNode)

	// Controlla le Istanze attive contattando il Service Registry

	// Continua finchè c'è almeno una istanza attiva
	result := JoinDHT(os.Args[1])
	for {
		if len(result) == 0 {
			result = JoinDHT(os.Args[1])
		} else {
			break
		}
	}
	fmt.Println(result)
	fmt.Println(len(result))

	// Se c'è solo un'istanza attiva, il nodo stesso crea il DHT Chord
	if len(result) == 1 {
		me = chord.Create(*addressPtr)
	} else {
		// Se c'è un'altra istanza attiva viene contattato un altro nodo random per fare la Join
		*joinPtr = result[rand.Intn(len(result))]
		for {
			if *joinPtr == *addressPtr {
				*joinPtr = result[rand.Intn(len(result))]
			} else {
				break
			}
		}
		*joinPtr = *joinPtr + ":4567"
		me, _ = chord.Join(*addressPtr, *joinPtr)
	}
	fmt.Printf("My address is: %s.\n", *addressPtr)
	fmt.Printf("Join address is: %s.\n", *joinPtr)
	fmt.Println("Chord Node Started Succesfully")
}

/*
func InitServiceRPC() *RPCservice {
	service := new(RPCservice)
	service.node = *me
	return service
}
*/
