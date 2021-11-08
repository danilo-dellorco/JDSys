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
	nodeRPC "progetto-sdcc/node/rpc"
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
	//testGetRPC()
	//testPutRPC()
	//testUpdateRPC()
	//testDeleteRPC()

	/*
		InitHealthyNode()
		InitChordDHT()
	*/
	mongoClient = mongo.InitLocalSystem()
	InitServiceRPC()
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
	port := ":4567"
	flag.Parse()

	// Ottiene l'indirizzo IP dell'host utilizzato nel VPC
	*addressPtr = GetOutboundIP().String()
	me = new(chord.ChordNode)

	// Controlla le Istanze attive contattando il Service Registry

	// Continua finchè c'è almeno una istanza attiva
waitLB:
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

	// Se c'è solo un'istanza attiva, se è il nodo stesso crea il DHT Chord, se non è lui
	// allora significa che non è ancora healthy per il LB e aspettiamo ad entrare nella rete
	if len(result) == 1 {
		if result[0] == *addressPtr {
			me = chord.Create(*addressPtr + port)
		} else {
			goto waitLB
		}
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
		me, _ = chord.Join(*addressPtr+port, *joinPtr+port)
	}
	fmt.Printf("My address is: %s.\n", *addressPtr)
	fmt.Printf("Join address is: %s.\n", *joinPtr)
	fmt.Printf("Port used: %s.\n", port)
	fmt.Println("Chord Node Started Succesfully")
}

/*
Inizializza il listener delle chiamate RPC. Và invocata dopo aver inizializzato sia Mongo che Chord
*/
func InitServiceRPC() {
	rpcServ := new(nodeRPC.RPCservice)
	rpcServ.Db = mongoClient
	//rpcServ.Node = me
	rpc.Register(rpcServ)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", utils.RPC_PORT)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}
