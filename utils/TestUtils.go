package utils

import (
	"fmt"
	"strings"
	"time"
)

var HL int = 80

func GetTimestamp() time.Time {
	return time.Now()
}

func FormatTime(t time.Time) string {
	//return t.Format("01-02-2006 15:04:05.000000000")
	return t.Format("15:04:05.000000000")
}

func PrintFormattedTimestamp() {
	fmt.Print("[" + FormatTime(GetTimestamp()) + "] ")
}

func getFormattedTimestamp() string {
	return "[" + FormatTime(GetTimestamp()) + "] "
}

func PrintTs(message string) {
	ts := getFormattedTimestamp()
	fmt.Print(ts + message + "\n")
}

func PrintHeaderL1(message string) {
	center := (HL-len(message))/2 - 2
	before := strings.Repeat("═", center) + "╣ "
	after := " ╠" + strings.Repeat("═", center)
	fmt.Print(before + message + after)
}

func PrintHeaderL2(message string) {
	fmt.Println("\n" + strings.Repeat("—", HL))
	PrintTs(message)
	fmt.Println(strings.Repeat("—", HL))
}

func PrintHeaderL3(message string) {
	fmt.Println("\n" + strings.Repeat("-", HL))
	PrintTs(message)
	fmt.Println(strings.Repeat("-", HL))
}

func PrintTailerL1() {
	fmt.Println(strings.Repeat("═", HL) + "\n")
}
