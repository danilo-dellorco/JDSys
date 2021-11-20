package mongo

import (
	"fmt"
	"time"
)

/*
Identifica un'entry di tipo {chiave,valore}, includendo il timestamp relativo
alla sua ultima modifica
*/
type MongoEntry struct {
	Key      string
	Value    string
	Timest   time.Time
	LastAcc  time.Time
	Conflict bool // rende piu efficiente il merge delle entry
}

/*
Stampa l'entry ed il relativo timestamp
*/
// TODO fare qualcosa per formattare l'entry
func (me *MongoEntry) Format() string {
	fmt.Print("{" + me.Key + ", " + me.Value + ", " + me.Timest.String() + "}")
	fmt.Printf(" %t\n", me.Conflict)
	return "TODO"
}
