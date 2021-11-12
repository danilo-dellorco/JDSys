package localsys

import (
	"fmt"
	"progetto-sdcc/node/localsys/communication"
	"progetto-sdcc/node/localsys/structures"
	"progetto-sdcc/utils"
)

/*
Inizializza il sistema di storage locale aprendo la connessione a MongoDB e lanciando le varie
routine per la ricezione degli update del DB dagli altri nodi e per la migrazione verso il cloud.
*/
func InitLocalSystem() structures.MongoClient {
	fmt.Println("Starting Mongo Local System...")
	client := structures.MongoClient{}
	client.OpenConnection()

	// Lancio della Goroutine tramite cui il nodo si mette in ascolto per la ricezione di DB updates:
	// 1. Tramite RPC da parte del Service Registry quando il nodo viene schedulato come "Terminating"
	// 2. Tramite invio periodico da parte di ogni nodo al suo successore
	go ListenUpdates(client)

	// Lancio della Goroutine tramite cui il nodo esporta periodicamente su S3 le entry accedute raramente
	go client.CheckRarelyAccessed()

	/*[TODO] Fare gestione di quando inviare gli aggiornamenti
	1) Ogni Tot Minuti per avere la consistenza finale*/

	return client
}

/*
Resta in ascolto per la ricezione di aggiornamenti del DB da altri nodi
*/
func ListenUpdates(cli structures.MongoClient) {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel)
	fmt.Println("Start receiving DB update from other nodes on port:", utils.UPDATES_PORT)
	for {
		received := <-fileChannel
		if received == "rcvd" {
			cli.UpdateCollection(utils.UPDATES_EXPORT_FILE, utils.UPDATES_RECEIVE_FILE)
			utils.ClearDir(utils.UPDATES_EXPORT_PATH)
			utils.ClearDir(utils.UPDATES_RECEIVE_PATH)
		}
	}
}

/*
Esporta il file CSV e lo invia al nodo remoto
*/
func SendUpdate(cli structures.MongoClient, address string) {
	file := utils.UPDATES_EXPORT_FILE
	cli.ExportCollection(file)
	communication.StartSender(file, address)
	utils.ClearDir(utils.UPDATES_EXPORT_PATH)
	utils.ClearDir(utils.UPDATES_RECEIVE_PATH)
}
