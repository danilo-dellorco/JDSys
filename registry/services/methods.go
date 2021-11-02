package services

import (
	"fmt"
)

// Struttura per il passaggio dei parametri alla RPC
type Args struct{}

// Pseudo-Interfaccia che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
// ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici!
type RingManagement int

// Metodo 1 dell'interfaccia
func (s *RingManagement) LockID(args *Args, reply *int) error {
	fmt.Printf("Stampa1")
	return nil
}

// Metodo 2 dell'interfaccia
func (s *RingManagement) UnlockID(args *Args, reply *int) error {
	fmt.Printf("Stampa2")
	return nil
}

func InitializeService() *RingManagement {
	service := new(RingManagement)
	return service
}
