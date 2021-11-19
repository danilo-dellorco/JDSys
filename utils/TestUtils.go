package utils

import (
	"fmt"
	"time"
)

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
	for i := 0; i <= len(message)+3; i++ {
		fmt.Print("*")
	}
	fmt.Println("\n*", message, "*")
	for i := 0; i <= len(message)+3; i++ {
		fmt.Print("*")
	}
}

func PrintHeaderL2(message string) {
	fmt.Println("\n----------------------------------------------------------------------------")
	PrintTs(message)
	fmt.Println("----------------------------------------------------------------------------")
}

func PrintTailerL2() {
	fmt.Println("----------------------------------------------------------------------------")
}

func PrintTailerL1(message string) {
	fmt.Printf(message + "\n")
	fmt.Printf("****************************************************************************\n")
}
