package localsys

import (
	"fmt"
	chord "progetto-sdcc/node/chord/net"
	"progetto-sdcc/node/localsys/communication"
	"progetto-sdcc/node/localsys/structures"
	"progetto-sdcc/utils"
)

//variabili globali per la realizzazione della consistenza finale
var Handler bool
var Round int

/*
Inizializza il sistema di storage locale aprendo la connessione a MongoDB e lanciando
i listener e le routine per la gestione degli updates.
*/
func InitLocalSystem(node *chord.ChordNode) structures.MongoClient {
	fmt.Println("Starting Mongo Local System...")
	client := structures.MongoClient{}
	client.OpenConnection()

	go ListenUpdateMessages(client)

	Handler = false
	Round = 0

	go ListenReconciliationMessages(client, node)

	fmt.Println("Mongo is Up & Running...")
	return client
}

/*
Resta in ascolto per messaggi di aggiornamento del database. Utilizzato per ricevere i DB dei nodi in terminazione
e le entry replicate.
*/
func ListenUpdateMessages(cli structures.MongoClient) {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel, "update")
	fmt.Println("Started Update Message listening Service...")
	for {
		received := <-fileChannel
		if received == "rcvd" {
			cli.MergeCollection(utils.UPDATES_EXPORT_FILE, utils.UPDATES_RECEIVE_FILE)
			utils.ClearDir(utils.UPDATES_EXPORT_PATH)
			utils.ClearDir(utils.UPDATES_RECEIVE_PATH)
		}
	}
}

/*
Resta in ascolto per la ricezione dei messaggi di riconciliazione. Ogni volta che si riceve un messaggio vengono
risolti i conflitti aggiornando il database
*/
func ListenReconciliationMessages(cli structures.MongoClient, node *chord.ChordNode) {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel, "reconciliation")
	fmt.Println("Started Reconciliation Message listening Service...")
	for {
		//si scrive sul canale per attivare la riconciliazione una volta ricevuto correttamente l'update dal predecessore
		received := <-fileChannel
		if received == "rcvd" {
			cli.ReconciliateCollection(utils.UPDATES_EXPORT_FILE, utils.UPDATES_RECEIVE_FILE)
			utils.ClearDir(utils.UPDATES_EXPORT_PATH)
			utils.ClearDir(utils.UPDATES_RECEIVE_PATH)

			fmt.Println("Mesa che schioppa al nodo")
			//nodo non ha successore, aspettiamo la ricostruzione della DHT Chord finchè non viene
			//completato l'aggiornamento dell'anello
		retry:
			if node.GetSuccessor() == nil {
				fmt.Println("Node hasn't a successor, wait for the reconstruction...")
				goto retry
			}

			//nodo effettua export del DB e lo invia al successore
			addr := node.GetSuccessor().GetIpAddr()
			fmt.Print("DB forwarded to successor:", addr, "\n\n")

			//solamente per il nodo che ha iniziato l'aggiornamento incrementiamo il contatore che ci permette
			//di interrompere dopo 2 giri non effettuando la SendCollectionMsg
			if Handler {
				Round++
				if Round == 2 {
					fmt.Println("Request returned to the node invoked by the registry two times, ring updates correctly")
					fmt.Print("========================================================\n\n\n")
					//ripristiniamo le variabili per le future riconciliazioni
					Handler = false
					Round = 0
				} else {
					SendCollectionMsg(cli, addr, "reconciliation")
				}
				//se il nodo è uno di quelli intermedi, si limita a propagare l'aggiornamento
			} else {
				SendCollectionMsg(cli, addr, "reconciliation")
			}
		}
	}
}

/*
Esporta il file CSV e lo invia al nodo remoto
*/
func SendCollectionMsg(cli structures.MongoClient, address string, mode string) {
	file := utils.UPDATES_EXPORT_FILE
	cli.ExportCollection(file)
	communication.StartSender(file, address, mode)
	utils.ClearDir(utils.UPDATES_EXPORT_PATH)
}
