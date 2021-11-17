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

	go ListenUpdates(client)

	fmt.Println("Mongo is Up & Running...")
	return client
}

/*
Resta in ascolto per la ricezione di aggiornamenti del DB da altri nodi per realizzare consistenza finale
*/
func ListenUpdates(cli structures.MongoClient) {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel)
	fmt.Println("Started Update Listening Service...")
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
Esporta il file CSV e lo invia al nodo remoto
*/
func SendUpdate(cli structures.MongoClient, address string) {
	file := utils.UPDATES_EXPORT_FILE
	cli.ExportCollection(file)
	communication.StartSender(file, address)
	utils.ClearDir(utils.UPDATES_EXPORT_PATH)
	utils.ClearDir(utils.UPDATES_RECEIVE_PATH)
}
