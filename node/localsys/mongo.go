package localsys

import (
	"fmt"
	"progetto-sdcc/node/localsys/structures"
)

/*
Inizializza il sistema di storage locale aprendo la connessione a MongoDB e lanciando
i listener e le routine per la gestione degli updates.
*/
func InitLocalSystem() structures.MongoClient {
	fmt.Println("Starting Mongo Local System...")
	client := structures.MongoClient{}
	client.OpenConnection()

	fmt.Println("Mongo is Up & Running...")
	return client
}
