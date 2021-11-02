package core

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"progetto-sdcc/node/mongo/communication"
	"progetto-sdcc/node/mongo/structures"
	"time"
)

func InitLocalSystem() {
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

	client.CloseConnection()
}

func ParseCSV(file string) []structures.MongoEntry {
	csvFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened CSV file")
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		fmt.Println(err)
	}

	var entryList []structures.MongoEntry
	i := 0
	for _, line := range csvLines {
		if i == 0 {
			i++
			continue
		}

		timeString := line[2]
		tVal, _ := time.Parse(time.RFC3339, timeString)
		entry := structures.MongoEntry{Key: line[0], Value: line[1], Timest: tVal, Conflict: false}
		entryList = append(entryList, entry)
	}
	return entryList
}

func mergeEntries(local []structures.MongoEntry, update []structures.MongoEntry) []structures.MongoEntry {
	var mergedEntries []structures.MongoEntry

	for i := 0; i < len(local); i++ {
		for j := 0; j < len(update); j++ {
			var latestEntry structures.MongoEntry
			if local[i].Key == update[j].Key {
				local[i].Conflict = true
				update[j].Conflict = true
				if local[i].Timest.After(update[j].Timest) {
					latestEntry = local[i]
				} else {
					latestEntry = update[j]
				}
				mergedEntries = append(mergedEntries, latestEntry)
			}
		}
		if !local[i].Conflict {
			mergedEntries = append(mergedEntries, local[i])
		}
	}
	for _, u := range update {
		if !u.Conflict {
			mergedEntries = append(mergedEntries, u)
		}
	}
	return mergedEntries
}

/**
* Resta in ascolto sulla ricezione di aggiornamenti del DB da altri nodi
**/
func ListenUpdates(cli structures.MongoClient) {
	fileChannel := make(chan string)
	go communication.StartReceiver(fileChannel)
	fmt.Println("diocane")
	for {
		received := <-fileChannel
		if received == "rcvd" {
			UpdateCollection(cli)
		}
	}
}

/**
* Si mette in attesa di ricevere aggiornamenti remoti. Ogni volta che si riceve un CSV viene aggiornato il database locale,
**/
func UpdateCollection(cli structures.MongoClient) {

	cli.ExportCollection("local/" + structures.LOCAL_CSV) // Export del LOCAL da mettere dopo la ricezione di update.csv, forse è meglio

	localList := ParseCSV("local/" + structures.LOCAL_CSV)
	updateList := ParseCSV("local/" + structures.UPDATE_CSV)
	mergedList := mergeEntries(localList, updateList)
	cli.Collection.Drop(context.TODO())
	for _, entry := range mergedList {
		cli.PutMongoEntry(entry)
	}
	cli.Collection.Find(context.TODO(), nil)
}

/**
* Esporta il file CSV e lo invia al nodo remoto
**/
func SendUpdate(cli structures.MongoClient) {
	file := "export/" + structures.UPDATE_CSV
	cli.ExportCollection(file)
	communication.StartSender(file)
}