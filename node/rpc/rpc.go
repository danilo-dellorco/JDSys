package net

import (
	"fmt"
	"log"
	"net/rpc"
	chord "progetto-sdcc/node/chord/net"
	mongo "progetto-sdcc/node/localsys"
	"progetto-sdcc/node/localsys/structures"
	"progetto-sdcc/utils"
)

//TODO testare il nuovo delete e verificare che fa tutto il giro dell'anello partendo dal nodo che deve gestire la risorsa

/*
Servizio RPC del nodo. Mantiene un riferimento al ChordNode ed al MongoClient
*/
type RPCservice struct {
	Node chord.ChordNode
	Db   structures.MongoClient
}

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
func (s *RPCservice) GetRPC(args *Args, reply *string) error {
	fmt.Println("GetRPC called!")

	fmt.Println("Checking value on local storage...")
	entry := s.Db.GetEntry(args.Key)
	if entry != nil {
		*reply = fmt.Sprintf("Key: %s\nValue: %s", entry.Key, entry.Value)
		return nil
	}

	fmt.Println("Key not found.")
	if s.Node.GetSuccessor().String() == "" {
		*reply = "Node hasn't a successor, wait for the reconstruction of the DHT and retry"
		return nil
	}

	fmt.Println("None.\nForwarding Get Request on DHT...")
	succ := s.Node.GetSuccessor().GetIpAddr()
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
func (s *RPCservice) PutRPC(args Args, reply *string) error {
	fmt.Println("PutRPC Called!")

	me := s.Node.GetIpAddress()
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)
	client, err := rpc.DialHTTP("tcp", utils.ParseAddrRPC(addr))
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Request send to:", utils.ParseAddrRPC(addr))
	client.Call("RPCservice.PutImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per aggiornare un'entry nello storage.
 1) Lookup per trovare il nodo che hosta la risorsa
 2) RPC effettiva di APPEND verso quel nodo chord
*/
func (s *RPCservice) AppendRPC(args Args, reply *string) error {
	fmt.Println("AppendRPC Called!")

	fmt.Println("Forwarding Append Request on DHT...")

	me := s.Node.GetIpAddress()
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
func (s *RPCservice) DeleteRPC(args Args, reply *string) error {
	fmt.Println("DeleteRPC called")

	me := s.Node.GetIpAddress()
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
func (s *RPCservice) GetImpl(args Args, reply *string) error {
	fmt.Println("Get request arrived")
	fmt.Println(args.Key)
	entry := s.Db.GetEntry(args.Key)
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
func (s *RPCservice) PutImpl(args Args, reply *string) error {
	fmt.Println("Put request arrived")
	arg1 := args.Key
	arg2 := args.Value
	err := s.Db.PutEntry(arg1, arg2)
	if err == nil {
		*reply = "Entry correctly inserted in the DB"
	} else {
		*reply = "Entry already in the DB"
		fmt.Println(*reply)
	}
	return nil
}

/*
Effettua l'APPEND. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (s *RPCservice) AppendImpl(args *Args, reply *string) error {
	fmt.Println("Append request arrived")
	arg1 := args.Key
	arg2 := args.Value
	fmt.Println("Arguments", arg1, arg2)
	err := s.Db.AppendValue(arg1, arg2)
	if err == nil {
		*reply = "Value correctly appended"
	} else {
		*reply = "Entry not found"
	}
	return nil
}

/*
Effettua il delete della risorsa sul nodo che deve gestirla.
Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (s *RPCservice) DeleteHandling(args *Args, reply *string) error {
	// Delete deve essere propagata a tutti i nodi, se il nodo che gestisce la precisa chiave non ha
	// un successore, non effettuiamo la cancellazione ma aspettiamo che venga ricostruita la DHT
	if s.Node.GetSuccessor().String() == "" {
		*reply = "Node hasn't a successor, wait for the reconstruction of the DHT and retry"
		return nil
	}

	// Nodo gestore ha correttamente un successore, procediamo con la delete sul DB locale
	fmt.Println("Deleting value on local storage...")
	err := s.Db.DeleteEntry(args.Key)
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
	next := s.Node.GetSuccessor().GetIpAddr()
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
func (s *RPCservice) DeleteReplicating(args *Args, reply *string) error {

	// La richiesta ha completato il giro dell'anello se è tornata al nodo che gestisce quella chiave
	if s.Node.GetIpAddress() == args.Handler {
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
	s.Db.DeleteEntry(args.Key)

	// Propaga la Delete al nodo successivo, la cancellazione sul nodo che gestisce la chiave
	// è già stata effettuata, per questo se i nodi successivi non hanno successore aspettiamo
	// la ricostruzione della DHT Chord finchè non viene completata la Delete!
retry:
	if s.Node.GetSuccessor().String() == "" {
		fmt.Println("Node hasn't a successor, wait for the reconstruction...")
		goto retry
	}
	next := s.Node.GetSuccessor().GetIpAddr()
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

func (s *RPCservice) TerminateInstanceRPC(args *Args, reply *string) error {
retry:
	if s.Node.GetSuccessor().String() == "" {
		fmt.Println("Node hasn't a successor, wait for the reconstruction of the DHT")
		goto retry
	}
	addr := s.Node.GetSuccessor().GetIpAddr()
	fmt.Println("Instance Scheduled to Terminating...")
	mongo.SendUpdate(s.Db, addr)
	*reply = "Instance Terminating"
	return nil
}

/*
Metodo invocato dal Service Registry quando le istanze EC2 devono procedere con lo scambio degli aggiornamenti
Effettua il trasferimento del proprio DB al nodo successore nella rete per realizzare la consistenza finale.
*/

func (s *RPCservice) ConsistencyHandlerRPC(args *Args, reply *string) error {
	fmt.Println("Final consistency requested by service registry...")

	if s.Node.GetSuccessor().String() == "" {
		*reply = "Node hasn't a successor, wait for the reconstruction of the DHT"
		return nil
	}

	//imposto il nodo da cui partirà l'aggiornamento dell'anello
	me := s.Node.GetIpAddress()
	args.Handler = me

	//nodo effettua export del DB e lo invia al successore
	addr := s.Node.GetSuccessor().GetIpAddr()
	fmt.Println("Sending DB export to my successor...")
	mongo.SendUpdate(s.Db, addr)

	//invoco esecuzione da parte del successore del trasferimento del DB
	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("Request forwarded to successor:", addr+utils.RPC_PORT)
	client.Call("RPCservice.ConsistencySuccessor", args, &reply)
	return nil
}

/*
Metodo invocato dal nodo random scelto dal Service Registry per raggiungere la consistenza finale
Ogni nodo invia il DB al successore, completato il giro, e quindi ritornati al nodo scelto dal registry,
se non si sono verificati aggiornamenti, tutti i dati saranno consistenti.
*/
func (s *RPCservice) ConsistencySuccessor(args *Args, reply *string) error {

	// La richiesta ha completato il giro dell'anello se è tornata al nodo che gestisce quella chiave
	if s.Node.GetIpAddress() == args.Handler {
		// campo usato come contatore per fare 2 giri nell'anello
		if !args.Deleted {
			args.Deleted = true
		} else {
			*reply = "Request returned to the node invoked by the registry two times, ring updates correctly"
			return nil
		}
	}

	// Se i nodi successivi non hanno successore aspettiamo la ricostruzione della DHT Chord
	//finchè non viene completato l'aggiornamento dell'anello
retry:
	if s.Node.GetSuccessor().String() == "" {
		fmt.Println("Node hasn't a successor, wait for the reconstruction...")
		goto retry
	}

	//nodo effettua export del DB e lo invia al successore
	addr := s.Node.GetSuccessor().GetIpAddr()
	fmt.Println("Sending DB export to my successor...")
	mongo.SendUpdate(s.Db, addr)

	//invoco esecuzione da parte del successore per continuare propagazione del DB nell'anello
	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("Request forwarded to successor:", addr+utils.RPC_PORT)
	client.Call("RPCservice.ConsistencySuccessor", args, &reply)
	return nil
}
