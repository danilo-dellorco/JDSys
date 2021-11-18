package main

import (
	"fmt"
	"log"
	"net/rpc"
	chord "progetto-sdcc/node/chord/net"
	"progetto-sdcc/utils"
)

// TODO dà un problema su Send Update e Receive forse perche servono porte diverse
// TODO Replicazione: Handler riceve la scrittura e la pro
// TODO testare filetransfer di riconciliazione e terminazione con le porte nuove eccetera

// TODO problema consistenza: il successore riceve l'export e parte la routine per il merge, il nodo
//avviando subito la RPC non da tempo al successore di completare il merge, per cui si effettua l'export
//del local DB per la Send presente nell'RPC ma il file viene buttato prima dell'invio dalla routine alla
//fine del merge
//--> intanto c'ho messo uno sleep di 3 secondi
//RIFLESSIONE: non possiamo farlo senza RPC, se uso SendCollectionMsg dopo il merge riesco a tx da nodo
//a successore ma non posso bloccare il giro, i parametri che passo alla funzione sono locali quindi non
//ho posso capire sui vari nodi chi è l'handler della riconciliazione!
//--> per forza RPC perchè devo passare argomenti tra i nodi, aggiungendo canale attiviamo su ogni nodo
//	  SendCollectionMsg al termine di merge + remove file, procedendo solo a questo punto con l'invio al succ

/*
Struttura che mantiene i parametri delle RPC
*/
type Args struct {
	Key     string
	Value   string
	Handler string
	Deleted bool
}

/*
Effettua la RPC per la Get di una Key.
 1) Si verifica se il nodo ha una copia della risorsa
 2) Lookup per trovare il nodo che hosta la risorsa
 3) RPC effettiva di GET verso quel nodo chord
*/
func (n *Node) GetRPC(args *Args, reply *string) error {
	fmt.Println("GetRPC called!")

	fmt.Println("Checking value on local storage...")
	entry := n.MongoClient.GetEntry(args.Key)
	if entry != nil {
		*reply = fmt.Sprintf("Key: %s\nValue: %s", entry.Key, entry.Value)
		return nil
	}

	fmt.Println("Key not found.")
	if n.ChordClient.GetSuccessor().String() == "" {
		*reply = "Node hasn't a successor, wait for the reconstruction of the DHT and retry"
		return nil
	}

	fmt.Println("None.\nForwarding Get Request on DHT...")
	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	addr, _ := chord.Lookup(utils.HashString(args.Key), succ+utils.CHORD_PORT)
	client, err := rpc.DialHTTP("tcp", utils.ParseAddrRPC(addr))
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Request send to:", utils.ParseAddrRPC(addr))
	client.Call("RPCservice.GetImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per inserire un'entry nello storage.
 1) Lookup per trovare il nodo che deve hostare la risorsa
 2) RPC effettiva di PUT verso quel nodo chord
*/
func (n *Node) PutRPC(args Args, reply *string) error {
	fmt.Println("PutRPC Called!")

	me := n.ChordClient.GetIpAddress()
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)
	client, err := rpc.DialHTTP("tcp", utils.ParseAddrRPC(addr))
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Request sent to:", utils.ParseAddrRPC(addr))
	client.Call("RPCservice.PutImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per aggiornare un'entry nello storage.
 1) Lookup per trovare il nodo che hosta la risorsa
 2) RPC effettiva di APPEND verso quel nodo chord
*/
func (n *Node) AppendRPC(args Args, reply *string) error {
	fmt.Println("AppendRPC Called!")

	fmt.Println("Forwarding Append Request on DHT...")

	me := n.ChordClient.GetIpAddress()
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)
	client, err := rpc.DialHTTP("tcp", utils.ParseAddrRPC(addr))
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Request send to:", utils.ParseAddrRPC(addr))
	client.Call("RPCservice.AppendImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per eliminare un'entry nello storage.
 1) Lookup per trovare il nodo che hosta la risorsa
 2) RPC effettiva di DELETE verso quel nodo chord
 3) La delete viene inoltrata su tutto l'anello
*/
func (n *Node) DeleteRPC(args Args, reply *string) error {
	fmt.Println("DeleteRPC called")

	me := n.ChordClient.GetIpAddress()
	handlerNode, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)
	args.Handler = utils.RemovePort(handlerNode)
	args.Deleted = false

	client, err := rpc.DialHTTP("tcp", utils.ParseAddrRPC(handlerNode))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("Delete request forwarded to handling node:", utils.ParseAddrRPC(handlerNode))
	client.Call("RPCservice.DeleteHandling", args, &reply)
	return nil
}

/*
Effettua il get. Scrive in reply la stringa contenente l'entry richiesta. Se l'entry
non è stata trovata restituisce un messaggio di errore.
*/
func (n *Node) GetImpl(args Args, reply *string) error {
	fmt.Println("Get request arrived")
	fmt.Println(args.Key)
	entry := n.MongoClient.GetEntry(args.Key)
	if entry == nil {
		*reply = "Entry not found"
	} else {
		*reply = fmt.Sprintf("Key: %s\nValue: %s", entry.Key, entry.Value)
	}
	return nil
}

/*
Effettua il PUT. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (n *Node) PutImpl(args Args, reply *string) error {
	fmt.Println("Put request arrived")
	arg1 := args.Key
	arg2 := args.Value
	err := n.MongoClient.PutEntry(arg1, arg2)
	ok := true
	if err == nil {
		*reply = "Entry correctly inserted in the DB"
	} else if err.Error() == "Update" {
		*reply = "Entry correctly updated"
	} else {
		*reply = err.Error()
		ok = false
	}

	// Se non ho avuto errori invo l'entry aggiunta al successore, che gestirà quindi una replica.
	if ok {
		next := n.ChordClient.GetSuccessor().GetIpAddr()
		n.MongoClient.ExportDocument(args.Key, utils.UPDATES_EXPORT_FILE)
		SendReplicationMsg(n, next, "replication")
	}
	return nil
}

/*
Effettua l'APPEND. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (n *Node) AppendImpl(args *Args, reply *string) error {
	fmt.Println("Append request arrived")
	arg1 := args.Key
	arg2 := args.Value
	err := n.MongoClient.AppendValue(arg1, arg2)
	if err == nil {
		*reply = "Value correctly appended"

		// Se non ho avuto errori invo l'entry aggiunta al successore, che gestirà quindi una replica.
		next := n.ChordClient.GetSuccessor().GetIpAddr()
		n.MongoClient.ExportDocument(args.Key, utils.UPDATES_EXPORT_FILE)
		SendReplicationMsg(n, next, "replication")
	} else {
		*reply = "Entry not found"
	}

	return nil
}

/*
Effettua il delete della risorsa sul nodo che deve gestirla.
Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (n *Node) DeleteHandling(args *Args, reply *string) error {
	// Delete deve essere propagata a tutti i nodi, se il nodo che gestisce la precisa chiave non ha
	// un successore, non effettuiamo la cancellazione ma aspettiamo che venga ricostruita la DHT
	if n.ChordClient.GetSuccessor().String() == "" {
		*reply = "Node hasn't a successor, wait for the reconstruction of the DHT and retry"
		return nil
	}

	// Nodo gestore ha correttamente un successore, procediamo con la delete sul DB locale
	fmt.Println("Deleting value on local storage...")
	err := n.MongoClient.DeleteEntry(args.Key)
	if err == nil {
		args.Deleted = true
	} else {
		// Entry non è presente nel DB del nodo gestore, quindi non esiste
		if err.Error() == "Entry Not Found" {
			*reply = "The key searched for delete not exist"
			return nil
		}
	}

	// Se l'entry esiste ed è stata cancellata, procediamo inoltrando la richiesta al nodo successore
	// così da eliminare tutte le repliche nell'anello
	next := n.ChordClient.GetSuccessor().GetIpAddr()
	client, err := rpc.DialHTTP("tcp", next+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("Delete request forwarded to replication node:", next+utils.RPC_PORT)
	client.Call("RPCservice.DeleteReplicating", args, &reply)
	return nil
}

/*
Effettua il delete della risorsa replicata.
Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (n *Node) DeleteReplicating(args *Args, reply *string) error {

	// La richiesta ha completato il giro dell'anello se è tornata al nodo che gestisce quella chiave
	if n.ChordClient.GetIpAddress() == args.Handler {
		fmt.Println("Request returned to the handler node")
		if args.Deleted {
			fmt.Println("Entry correctly removed from every node!")
			*reply = "Entry succesfully deleted"
		} else {
			*reply = "Entry to delete not found"
		}
		return nil
	}

	// Cancella la richiesta sul db locale
	fmt.Println("Deleting replicated value on local storage...")
	n.MongoClient.DeleteEntry(args.Key)

	// Propaga la Delete al nodo successivo, la cancellazione sul nodo che gestisce la chiave
	// è già stata effettuata, per questo se i nodi successivi non hanno successore aspettiamo
	// la ricostruzione della DHT Chord finchè non viene completata la Delete!
retry:
	if n.ChordClient.GetSuccessor().String() == "" {
		fmt.Println("Node hasn't a successor, wait for the reconstruction...")
		goto retry
	}
	next := n.ChordClient.GetSuccessor().GetIpAddr()
	client, err := rpc.DialHTTP("tcp", next+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("Delete request forwarded to:", next+utils.RPC_PORT)
	client.Call("RPCservice.DeleteReplicating", args, &reply)
	return nil
}

/*
Metodo invocato dal Service Registry quando l'istanza EC2 viene schedulata per la terminazione
Effettua il trasferimento del proprio DB al nodo successore nella rete per garantire replicazione dei dati.
Inviamo tutto il DB e non solo le entry gestite dal preciso nodo così abbiamo la possibilità di
aggiornare altri dati obsoleti mantenuti dal successore
*/

func (n *Node) TerminateInstanceRPC(args *Args, reply *string) error {
retry:
	if n.ChordClient.GetSuccessor().String() == "" {
		fmt.Println("Node hasn't a successor, wait for the reconstruction of the DHT")
		goto retry
	}
	addr := n.ChordClient.GetSuccessor().GetIpAddr()
	fmt.Println("Instance Scheduled to Terminating...")
	SendReplicationMsg(n, addr, "update")
	*reply = "Instance Terminating"
	return nil
}

/*
Metodo invocato dal Service Registry quando le istanze EC2 devono procedere con lo scambio degli aggiornamenti
Effettua il trasferimento del proprio DB al nodo successore nella rete per realizzare la consistenza finale.
*/

func (n *Node) ConsistencyHandlerRPC(args *Args, reply *string) error {
	fmt.Println("\n\n========================================================")
	fmt.Println("Final consistency requested by service registry...")

	if n.ChordClient.GetSuccessor().String() == "" {
		*reply = "Node hasn't a successor, abort and wait for the reconstruction of the DHT."
		fmt.Println(*reply)
		return nil
	}

	//imposto il nodo corrente come gestore dell'aggiornamento dell'anello, così da incrementare solo
	//per lui il contatore che permette l'interruzione dopo 2 giri
	n.Handler = true

	//nodo effettua export del DB e lo invia al successore
	addr := n.ChordClient.GetSuccessor().GetIpAddr()
	SendReplicationMsg(n, addr, "reconciliation")

	//invoco esecuzione da parte del successore del trasferimento del DB
	//client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	//if err != nil {
	//	log.Fatal("dialing:", err)
	//}

	//non forwardiamo immediatamente la richiesta al successore, così gli diamo il tempo di fare il
	//merge dei DB prima di gestire l'RPC, che richiederà il suo export da inviare al suo successore
	//--> senza l'attesa il nodo fa l'export del DB, ma la routine che riceve l'export da questo fa il merge e butta l'export locale, quindi non si trova il file creato per l'invio!

	//fmt.Print("Request forwarded to successor:", addr+utils.RPC_PORT, "\n\n\n")
	//time.Sleep(3 * time.Second)
	//client.Call("RPCservice.ConsistencySuccessor", args, &reply)
	return nil
}

/*
Metodo invocato dal nodo random scelto dal Service Registry per raggiungere la consistenza finale
Ogni nodo invia il DB al successore, completato il giro, e quindi ritornati al nodo scelto dal registry,
se non si sono verificati aggiornamenti, tutti i dati saranno consistenti.
*/
/*
func (n *Node) ConsistencySuccessor(args *Args, reply *string) error {
	fmt.Println("Incoming request for final consistency, start update successor...")
	// La richiesta ha completato il giro dell'anello se è tornata al nodo che gestisce quella chiave
	if n.ChordClient.GetIpAddress() == args.Handler {
		args.Round++
		if args.Round == 2 {
			*reply = "Request returned to the node invoked by the registry two times, ring updates correctly"
			fmt.Println(*reply)
			fmt.Print("========================================================\n\n\n")
			return nil
		}
	}

	// Se i nodi successivi non hanno successore aspettiamo la ricostruzione della DHT Chord
	//finchè non viene completato l'aggiornamento dell'anello
retry:
	if n.ChordClient.GetSuccessor().String() == "" {
		fmt.Println("Node hasn't a successor, wait for the reconstruction...")
		goto retry
	}

	//nodo effettua export del DB e lo invia al successore
	addr := n.ChordClient.GetSuccessor().GetIpAddr()
	n.SendReplicationMsg(addr, "reconciliation")

	//invoco esecuzione da parte del successore per continuare propagazione del DB nell'anello
	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	//non forwardiamo immediatamente la richiesta al successore, così gli diamo il tempo di fare il
	//merge dei DB prima di gestire l'RPC, che richiederà il suo export da inviare al suo successore
	//--> senza l'attesa il nodo fa l'export del DB, ma la routine che riceve l'export da questo fa il merge e butta l'export locale, quindi non si trova il file creato per l'invio!
	fmt.Print("Request forwarded to successor:", addr+utils.RPC_PORT, "\n\n\n")
	//time.Sleep(3 * time.Second)
	client.Call("RPCservice.ConsistencySuccessor", args, &reply)
	return nil
}
*/
