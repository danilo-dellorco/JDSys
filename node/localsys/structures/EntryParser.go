package structures

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

/*
Ottiene una lista di Entry partendo da un file CSV
*/
func ParseCSV(file string) []MongoEntry {
	fmt.Println("Parsing CSV:", file)
	csvFile, err := os.Open(file)
	if err != nil {
		fmt.Println("ParseCSV Error:", err)
	}

	csvr := csv.NewReader(csvFile)

	// Disabilitiamo il check della lunghezza dei record
	// così non va in errore ReadAll se c'è riga vuota
	csvr.FieldsPerRecord = -1
	csvLines, err := csvr.ReadAll()
	if err != nil {
		fmt.Println("ReadCSV Error:", err)
	}

	var entryList []MongoEntry
	i := 0
	for _, line := range csvLines {
		if i == 0 {
			i++
			continue
		}
		//trascuriamo record con meno campi del previsto
		if len(line) < 4 {
			continue
		}
		timeString := line[2]
		tVal, _ := time.Parse(time.RFC3339, timeString)
		accessString := line[3]
		aVal, _ := time.Parse(time.RFC3339, accessString)
		entry := MongoEntry{Key: line[0], Value: line[1], Timest: tVal, LastAcc: aVal, Conflict: false}
		entryList = append(entryList, entry)
	}
	defer csvFile.Close()
	fmt.Println("CSV Parsed correctly")
	return entryList
}

/*
Unisce le Entry tenendo in caso di conflitti sempre quella piu recente
*/
func MergeEntries(local []MongoEntry, update []MongoEntry) []MongoEntry {
	fmt.Println("Merging Database Entries...")

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

				// Appendo l'entry con conflict a false.
				temp := latestEntry
				temp.Conflict = false
				mergedEntries = append(mergedEntries, temp)
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
	fmt.Println("Entries:")
	return mergedEntries
}

/*
Risolve i conflitti secondo Last Write Wins
*/
func ReconciliateEntries(local []MongoEntry, update []MongoEntry) []MongoEntry {
	fmt.Println("Merging Database Entries...")

	var reconEntries []MongoEntry

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

				// Appendo l'entry con conflict a false.
				temp := latestEntry
				temp.Conflict = false
				reconEntries = append(reconEntries, temp)
			}
		}
		if !local[i].Conflict {
			reconEntries = append(reconEntries, local[i])
		}
	}
	for _, u := range local {
		if !u.Conflict {
			reconEntries = append(reconEntries, u)
		}
	}
	fmt.Println("Entries:")
	return reconEntries
}
