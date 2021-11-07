package utils

import (
	"fmt"
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
}
