package main

import (
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
	"progetto-sdcc/node/localsys/communication"
	"progetto-sdcc/node/localsys/structures"
	"progetto-sdcc/utils"
	"time"
)

type EmptyArgs struct{}

type Node struct {
	MongoClient structures.MongoClient
	ChordClient *chord.ChordNode

	//variabili per la realizzazione della consistenza finale
	Handler bool
	Round   int
}

func main() {
	//NodeLocalSetup()
	node := new(Node)
	node.nodeSetup()

Loop:
	for {
		var cmd string
		_, err := fmt.Scan(&cmd)
		switch {
		case cmd == "print":
			//stampa successore e predecessore
			fmt.Printf("%s", node.ChordClient.String())
		case cmd == "fingers":
			//stampa la finger table
			fmt.Printf("%s", node.ChordClient.ShowFingers())
		case cmd == "succ":
			//stampa la lista di successori
			fmt.Printf("%s", node.ChordClient.ShowSucc())
		case err == io.EOF:
			break Loop
		}

	}
	node.ChordClient.Finalize()
	select {}
}

/*
Gestisce gli hearthbeat del Load Balancer ed i messaggi di Terminazione dal Service Registry
*/
func lb_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "SDCC Distributed Key-Value Storage")
}

/*
Inizializza un listener sulla porta 8888, su cui il Nodo riceve gli HeartBeat del Load Balancer,
ed i segnali di terminazione dal service registry.
*/
func StartHeartBeatListener() {
	fmt.Println("Start Listening Heartbeats from LB on port:", utils.HEARTBEAT_PORT)
	http.HandleFunc("/", lb_handler)
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
func HttpConnect(registryAddr string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", registryAddr+utils.REGISTRY_PORT)
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return client, err
}

/*
Permette al nodo di inserirsi nell'anello chord contattando il server specificato
*/
func JoinDHT(registryAddr string) []string {
	args := EmptyArgs{}
	var reply []string

	client, _ := HttpConnect(registryAddr)
	err := client.Call("DHThandler.JoinRing", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	return reply
}

/*
Permette al nodo di essere rilevato come Healthy Instance dal Load Balancer e configura il DB locale
*/
func (n *Node) initHealthyNode() {

	// Configura il sistema di storage locale
	n.MongoClient = mongo.InitLocalSystem()

	// Inizia a ricevere gli HeartBeat dal LB
	go StartHeartBeatListener()

	// Attende di diventare healthy per il Load Balancer
	fmt.Println("Waiting for ELB Health Checking...")
	time.Sleep(utils.NODE_HEALTHY_TIME)
	fmt.Println("EC2 Node Up & Running!")
}

/*
Permette al nodo di entrare a far parte della DHT Chord in base alle informazioni ottenute dal Service Registry.
Inizia anche due routine per aggiornamento periodico delle FT del nodo stesso e degli altri nodi della rete
*/
func (n *Node) initChordDHT() {
	fmt.Println("Initializing Chord DHT...")

	// Setup dei Flags
	addressPtr := flag.String("addr", "", "the port you will listen on for incomming messages")
	joinPtr := flag.String("join", "", "an address of a server in the Chord network to join to")
	flag.Parse()

	// Ottiene l'indirizzo IP dell'host utilizzato nel VPC
	*addressPtr = GetOutboundIP().String()
	n.ChordClient = new(chord.ChordNode)

	// Controlla le istanze attive contattando il Service Registry per entrare nella rete
waitLB:
	result := JoinDHT(os.Args[1])
	for {
		if len(result) == 0 {
			result = JoinDHT(os.Args[1])
		} else {
			break
		}
	}

	// Unica istanza attiva, se è il nodo stesso crea la DHT Chord, se non è lui
	// allora significa che non è ancora healthy per il LB e aspettiamo ad entrare nella rete
	if len(result) == 1 {
		if result[0] == *addressPtr {
			n.ChordClient = chord.Create(*addressPtr + utils.CHORD_PORT)
		} else {
			goto waitLB
		}
	} else {
		// Se c'è più di un'istanza attiva viene contattato un altro nodo random per fare la Join
		*joinPtr = result[rand.Intn(len(result))]
		for {
			if *joinPtr == *addressPtr {
				*joinPtr = result[rand.Intn(len(result))]
			} else {
				break
			}
		}
		n.ChordClient, _ = chord.Join(*addressPtr+utils.CHORD_PORT, *joinPtr+utils.CHORD_PORT)
	}
	fmt.Printf("My address is: %s.\n", *addressPtr)
	fmt.Printf("Join address is: %s.\n", *joinPtr)
	fmt.Println("Chord Node Started Succesfully!")
}

/*
Inizializza il listener delle chiamate RPC per il funzionamento del sistema di storage distribuito.
Và invocata dopo aver inizializzato sia MongoDB che la DHT Chord in modo da poter gestire correttamente la comunicazione
tra i nodi del sistema.
*/
func (n *Node) initRPCService() {
	rpc.Register(n)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", utils.RPC_PORT)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	fmt.Println("RPC Service Started...")
	fmt.Println("Start Serving RPC request on port:", utils.RPC_PORT)
	go http.Serve(l, nil)
}

/*
Esegue tutte le attività per rendere il nodo UP & Running
*/
func (n *Node) nodeSetup() {
	n.initHealthyNode()
	n.initChordDHT()
	n.initRPCService()

	go n.ListenUpdateMessages()

	n.Handler = false
	n.Round = 0
	go n.ListenReconciliationMessages()
}

/*
Resta in ascolto per messaggi di aggiornamento del database. Utilizzato per ricevere i DB dei nodi in terminazione
e le entry replicate.
*/
func (n *Node) ListenUpdateMessages() {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel, "update")
	fmt.Println("Started Update Message listening Service...")
	for {
		received := <-fileChannel
		if received == "rcvd" {
			n.MongoClient.MergeCollection(utils.UPDATES_EXPORT_FILE, utils.UPDATES_RECEIVE_FILE)
			utils.ClearDir(utils.UPDATES_EXPORT_PATH)
			utils.ClearDir(utils.UPDATES_RECEIVE_PATH)
		}
	}
}

/*
Resta in ascolto per la ricezione dei messaggi di riconciliazione. Ogni volta che si riceve un messaggio vengono
risolti i conflitti aggiornando il database
*/
func (n *Node) ListenReconciliationMessages() {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel, "reconciliation")
	fmt.Println("Started Reconciliation Message listening Service...")
	for {
		//si scrive sul canale per attivare la riconciliazione una volta ricevuto correttamente l'update dal predecessore
		received := <-fileChannel
		if received == "rcvd" {
			n.MongoClient.ReconciliateCollection(utils.UPDATES_EXPORT_FILE, utils.UPDATES_RECEIVE_FILE)
			utils.ClearDir(utils.UPDATES_EXPORT_PATH)
			utils.ClearDir(utils.UPDATES_RECEIVE_PATH)

			fmt.Println("Mesa che schioppa al nodo")
			//nodo non ha successore, aspettiamo la ricostruzione della DHT Chord finchè non viene
			//completato l'aggiornamento dell'anello
		retry:
			if n.ChordClient.GetSuccessor().String() == "" {
				fmt.Println("Node hasn't a successor, wait for the reconstruction...")
				goto retry
			}

			//nodo effettua export del DB e lo invia al successore
			addr := n.ChordClient.GetSuccessor().GetIpAddr()
			fmt.Print("DB forwarded to successor:", addr, "\n\n")

			//solamente per il nodo che ha iniziato l'aggiornamento incrementiamo il contatore che ci permette
			//di interrompere dopo 2 giri non effettuando la SendCollectionMsg
			if n.Handler {
				n.Round++
				if n.Round == 2 {
					fmt.Println("Request returned to the node invoked by the registry two times, ring updates correctly")
					fmt.Print("========================================================\n\n\n")
					//ripristiniamo le variabili per le future riconciliazioni
					n.Handler = false
					n.Round = 0
				} else {
					n.SendReplicationMsg(addr, "reconciliation")
				}
				//se il nodo è uno di quelli intermedi, si limita a propagare l'aggiornamento
			} else {
				n.SendReplicationMsg(addr, "reconciliation")
			}
		}
	}
}

/*
Esporta il file CSV e lo invia al nodo remoto
*/
func (n *Node) SendReplicationMsg(address string, mode string) {
	file := utils.UPDATES_EXPORT_FILE
	n.MongoClient.ExportCollection(file)
	communication.StartSender(file, address, mode)
	utils.ClearDir(utils.UPDATES_EXPORT_PATH)
}
