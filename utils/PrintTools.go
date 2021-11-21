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

func StringInBox(message string) string {
	top := "+" + strings.Repeat("—", len(message)+2) + "+\n"
	middle := "| " + message + " |\n"
	bottom := top

	return top + middle + bottom
}

func StringInBoxL2(msg1 string, msg2 string) string {
	var lenght int
	var diff1 int
	var diff2 int
	if len(msg1) >= len(msg2) {
		lenght = len(msg1)
		diff1 = 0
		diff2 = lenght - len(msg2)
	} else {
		lenght = len(msg2)
		diff2 = 0
		diff1 = lenght - len(msg1)
	}
	top := "+" + strings.Repeat("—", lenght+2) + "+\n"
	middle1 := "| " + msg1 + strings.Repeat(" ", diff1) + " |\n"
	middle2 := "| " + msg2 + strings.Repeat(" ", diff2) + " |\n"
	bottom := top

	return top + middle1 + middle2 + bottom
}
