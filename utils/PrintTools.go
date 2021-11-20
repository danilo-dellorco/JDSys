package utils

import (
	"fmt"
	"strings"
	"time"
)

var HL int = 80

/*
Ritorna il valore attuale del tempo
*/
func GetTimestamp() time.Time {
	return time.Now()
}

/*
Ritorna una stringa con il valore del tempo formattato
*/
func FormatTime(t time.Time) string {
	return t.Format("15:04:05.0000")
}

/*
Stampa un timestamp
*/
func PrintFormattedTimestamp() {
	fmt.Print("[" + FormatTime(GetTimestamp()) + "] ")
}

/*
Stampa una stringa, includendo un timestamp formattato
*/
func PrintTs(message string) {
	ts := "[" + FormatTime(GetTimestamp()) + "] "
	fmt.Print(ts + message + "\n")
}

/*
Stampa un messaggio formattandolo come Header di Livello 1
*/
func PrintHeaderL1(message string) {
	center := (HL-len(message))/2 - 2
	before := strings.Repeat("═", center) + "╣ "
	after := " ╠" + strings.Repeat("═", center)
	fmt.Print(before + message + after)
}

/*
Stampa un messaggio formattandolo come Header di Livello 2
*/
func PrintHeaderL2(message string) {
	fmt.Println("\n" + strings.Repeat("—", HL))
	PrintTs(message)
	fmt.Println(strings.Repeat("—", HL))
}

/*
Stampa un messaggio formattandolo come Header di Livello 3
*/
func PrintHeaderL3(message string) {
	fmt.Println("\n" + strings.Repeat("-", HL))
	PrintTs(message)
	fmt.Println(strings.Repeat("-", HL))
}

/*
Stampa una stringa per chiudere l'Header di Livello 1
*/
func PrintTailerL1() {
	fmt.Println(strings.Repeat("═", HL) + "\n")
}
