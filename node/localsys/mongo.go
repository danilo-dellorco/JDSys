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

	// Lancio della Goroutine che permette al nodo di restare in attesa perenne
	go ListenUpdates(client)

	/*[TODO] Fare gestione di quando inviare gli aggiornamenti
	1) Ogni Tot Minuti per avere la consistenza finale
	2) Quando un nodo ESCE dall'anello deve inviare il suo db per fare merge
	==> SendUpdate non và chiamata nel main di default, ma invocata in risposta agli eventi (1) e (2)
	SendUpdate(client)
	*/

	// ***************** TEST *********************
	client.PutEntry("TestKey", "TestValue")
	client.PutEntry("TestKey1", "TestValue1")
	client.PutEntry("TestKey2", "TestValue2")
	client.PutEntry("TestKey3", "TestValue3")
	/*
		//client.CheckRarelyAccessed()
		client.GetEntry("TestKey")
		client.GetEntry("TestKey1")
		client.GetEntry("TestKey2")
		client.GetEntry("TestKey3")
		//client.CloseConnection()
	*/
	// ********************************************
	return client
}

/*
Resta in ascolto per la ricezione di aggiornamenti del DB da altri nodi
*/
func ListenUpdates(cli structures.MongoClient) {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel)
	fmt.Println("Start Receiving DB update from other nodes on port:", utils.UPDATES_PORT)
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
