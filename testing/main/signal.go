package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		for {
			sig := <-sigs
			if sig == os.Interrupt {
				printSignal()
			}
		}
	}()

	for {

	}
}

func printSignal() {
	fmt.Println("CatturatoSegnale")
	for {
		fmt.Printf("=")
	}
}
