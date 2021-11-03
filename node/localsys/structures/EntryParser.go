package structures

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

/**
* Unisce le Entry tenendo in caso di conflitti sempre quella piu recente
**/
func MergeEntries(local []MongoEntry, update []MongoEntry) []MongoEntry {
	var mergedEntries []MongoEntry

	for i := 0; i < len(local); i++ {
		for j := 0; j < len(update); j++ {
			var latestEntry MongoEntry
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
* Ottiene una lista di Entry partendo da un file CSV
**/
func ParseCSV(file string) []MongoEntry {
	csvFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		fmt.Println(err)
	}

	var entryList []MongoEntry
	i := 0
	for _, line := range csvLines {
		if i == 0 {
			i++
			continue
		}
		timeString := line[2]
		tVal, _ := time.Parse(time.RFC3339, timeString)
		accessString := line[3]
		aVal, _ := time.Parse(time.RFC3339, accessString)
		entry := MongoEntry{Key: line[0], Value: line[1], Timest: tVal, LastAcc: aVal, Conflict: false}
		entryList = append(entryList, entry)
	}
	return entryList
}
