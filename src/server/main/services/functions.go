package services

import (
	"fmt"
)

//struttura per il passaggio dei parametri nella RPC
type Args struct {
	A, B int
}

//"interfaccia" che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
//ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici!
type ServizioDiProva int

// Metodo 1 dell'interfaccia
func (s *ServizioDiProva) Stampa1(args *Args, reply *int) error{
	fmt.Printf("Stampa1")
	return nil
}

// Metodo 2 dell'interfaccia
func (s *ServizioDiProva) Stampa2(args *Args, reply *int) error{
	fmt.Printf("Stampa2")
	return nil
}

// Metodo che lista tutti i Metodi Remoti disponibili
func (s *ServizioDiProva) ListMethods(args *Args, reply *int) error{
	fmt.Printf("1) Stampa1\n")
	fmt.Printf("2) Stampa2\n")
	return nil
}

func InitializeService() *ServizioDiProva{
	service := new (ServizioDiProva)
	return service
}