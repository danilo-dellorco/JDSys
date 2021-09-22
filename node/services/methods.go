package services

import (
	"fmt"
	"log"
	"net/http"
)

func home_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")
}

func pica_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Picapage")
}

func ListenHttpConnection() {
	//gestisce richieste per un preciso percorso tramite l'handler specificato
	//se il percorso non è opportunamente gestito, viene usato l'handler della root
	//es. http://IP:80/ennio
	// NomeDNS/pica renderizza un'altra pagina
	http.HandleFunc("/", home_handler)
	http.HandleFunc("/pica", pica_handler)
	log.Fatal(http.ListenAndServe(":80", nil))
}

// Struttura per il passaggio dei parametri alla RPC
type Args struct {
	A, B int
}

// Pseudo-Interfaccia che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
// ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici!
type ServizioDiProva int

// Metodo 1 dell'interfaccia
func (s *ServizioDiProva) Stampa1(args *Args, reply *int) error {
	fmt.Printf("Stampa1")
	return nil
}

// Metodo 2 dell'interfaccia
func (s *ServizioDiProva) Stampa2(args *Args, reply *int) error {
	fmt.Printf("Stampa2")
	return nil
}

// Metodo 3 dell'interfaccia che lista tutti i Metodi Remoti disponibili
func (s *ServizioDiProva) ListMethods(args *Args, reply *string) error {
	*reply = "============= METHODS LIST =============\n" +
		"1) Metodo1\n" +
		"2) Metodo2\n"
	return nil
}

func InitializeService() *ServizioDiProva {
	service := new(ServizioDiProva)
	return service
}
