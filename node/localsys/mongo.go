package localsys

import (
	"fmt"
	"progetto-sdcc/node/localsys/communication"
	"progetto-sdcc/node/localsys/structures"
	"progetto-sdcc/utils"
)

/*
Inizializza il sistema di storage locale aprendo la connessione a MongoDB e lanciando
i listener e le routine per la gestione degli updates.
*/
func InitLocalSystem() structures.MongoClient {
	fmt.Println("Starting Mongo Local System...")
	client := structures.MongoClient{}
	client.OpenConnection()

	go ListenUpdateMessages(client)
	go ListenReconciliationMessages(client)

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
func ListenReconciliationMessages(cli structures.MongoClient) {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel, "reconciliation")
	fmt.Println("Started Reconciliation Message listening Service...")
	for {
		received := <-fileChannel
		if received == "rcvd" {
			cli.ReconciliateCollection(utils.UPDATES_EXPORT_FILE, utils.UPDATES_RECEIVE_FILE)
			utils.ClearDir(utils.UPDATES_EXPORT_PATH)
			utils.ClearDir(utils.UPDATES_RECEIVE_PATH)
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
