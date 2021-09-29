package structures

import (
	"fmt"
	"time"
)

/**
* Identifica un'entry di tipo {chiave,valore}, includendo il timestamp relativo
* alla sua ultima modifica
**/
type MongoEntry struct {
	Key      string
	Value    string
	Timest   time.Time
	Conflict bool // rende piu efficiente il merge delle entry
}

/**
* Stampa l'entry ed il relativo timestamp
**/
func (me *MongoEntry) print() {
	fmt.Print("{" + me.Key + ", " + me.Value + ", " + me.Timest.String() + "}")
	fmt.Printf(" %t\n", me.Conflict)
}
