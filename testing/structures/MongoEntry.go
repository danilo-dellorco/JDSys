package structures

import (
	"fmt"
	"time"
)

type MongoEntry struct {
	Key      string
	Value    string
	Timest   time.Time
	Conflict bool // rende piu efficiente il merge delle entry
}

func (me *MongoEntry) print() {
	fmt.Print("{" + me.Key + ", " + me.Value + ", " + me.Timest.String() + "}")
	fmt.Printf(" %t\n", me.Conflict)
}
