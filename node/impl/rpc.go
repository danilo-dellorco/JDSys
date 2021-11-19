package impl

import (
	"fmt"
	chord "progetto-sdcc/node/chord/api"
	mongo "progetto-sdcc/node/mongo/api"
	"progetto-sdcc/utils"
	"time"
)

type Node struct {
	MongoClient mongo.MongoInstance
	ChordClient *chord.ChordNode

	// Variabili per la realizzazione della consistenza finale
	Handler bool
	Round   int
}

// TODO testare filetransfer di riconciliazione e terminazione con le porte nuove eccetera
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
	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	if succ == "" {
		*reply = "Node hasn't a successor, wait for the reconstruction of the DHT and retry"
		return nil
	}

	fmt.Println("None.\nForwarding Get Request on DHT...")
	addr, _ := chord.Lookup(utils.HashString(args.Key), succ+utils.CHORD_PORT)
	client, _ := utils.HttpConnect(utils.RemovePort(addr), utils.RPC_PORT)
	fmt.Println("Request send to:", utils.ParseAddrRPC(addr))
	client.Call("Node.GetImpl", args, &reply)
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
	client, _ := utils.HttpConnect(utils.RemovePort(addr), utils.RPC_PORT)
	fmt.Println("Request sent to:", utils.ParseAddrRPC(addr))
	client.Call("Node.PutImpl", args, &reply)
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
	client, _ := utils.HttpConnect(utils.RemovePort(addr), utils.RPC_PORT)

	fmt.Println("Request send to:", utils.ParseAddrRPC(addr))
	client.Call("Node.AppendImpl", args, &reply)
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

	client, _ := utils.HttpConnect(utils.RemovePort(handlerNode), utils.RPC_PORT)
	fmt.Println("Delete request forwarded to handling node:", utils.ParseAddrRPC(handlerNode))
	client.Call("Node.DeleteHandling", args, &reply)
	return nil
}

/*
Metodo invocato dal Service Registry quando le istanze EC2 devono procedere con lo scambio degli aggiornamenti
Effettua il trasferimento del proprio DB al nodo successore nella rete per realizzare la consistenza finale.
*/
func (n *Node) ConsistencyHandlerRPC(args *Args, reply *string) error {
	fmt.Println("\n\n========================================================")
	fmt.Println("Final consistency requested by service registry...")

	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	if succ == "" {
		*reply = "Node hasn't a successor, abort and wait for the reconstruction of the DHT."
		fmt.Println(*reply)
		return nil
	}

	// Imposto il nodo corrente come gestore dell'aggiornamento dell'anello, così da incrementare solo
	// per lui il contatore che permette l'interruzione dopo 2 giri
	n.Handler = true

	// Effettuo l' export del DB e lo invio al successore
	SendReplicationMsg(n, succ, "reconciliation")
	return nil
}

/*
Metodo invocato dal Service Registry quando l'istanza EC2 viene schedulata per la terminazione
Effettua il trasferimento del proprio DB al nodo successore nella rete per garantire replicazione dei dati.
Inviamo tutto il DB e non solo le entry gestite dal preciso nodo così abbiamo la possibilità di
aggiornare altri dati obsoleti mantenuti dal successore.
*/

func (n *Node) TerminateInstanceRPC(args *Args, reply *string) error {
	fmt.Println("Instance Scheduled to Terminating...")
retry:
	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	if succ == "" {
		fmt.Println("Node hasn't a successor, wait for the reconstruction of the DHT")
		time.Sleep(2 * time.Second)
		goto retry
	}
	SendReplicationMsg(n, succ, "update")
	*reply = "Instance Terminating"
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

	// Se non ho avuto errori, se è presente il successore inviamo l'entry per fargli gestire una replica.
	if ok {
		succ := n.ChordClient.GetSuccessor().GetIpAddr()
		if succ == "" {
			fmt.Println("Node hasn't a successor, data will be replicated later")
			return nil
		}
		n.MongoClient.ExportDocument(args.Key, utils.UPDATES_EXPORT_FILE)
		SendReplicationMsg(n, succ, "replication")
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
		succ := n.ChordClient.GetSuccessor().GetIpAddr()
		n.MongoClient.ExportDocument(args.Key, utils.UPDATES_EXPORT_FILE)
		SendReplicationMsg(n, succ, "replication")
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
	// Se non è presente, il nodo potrebbe essere da solo, o un eventuale successore ancora non identificato
	// verrà poi aggiornato successivamente tramite la riconciliazione
	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	if succ == "" {
		*reply = "Entry successfully deleted"
		return nil
	}
	client, _ := utils.HttpConnect(succ, utils.RPC_PORT)
	fmt.Println("Delete request forwarded to replication node:", succ+utils.RPC_PORT)
	client.Call("Node.DeleteReplicating", args, &reply)
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
	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	if succ == "" {
		fmt.Println("Node hasn't a successor, wait for the reconstruction...")
		time.Sleep(2 * time.Second)
		goto retry
	}
	client, _ := utils.HttpConnect(succ, utils.RPC_PORT)
	fmt.Println("Delete request forwarded to:", succ+utils.RPC_PORT)
	client.Call("Node.DeleteReplicating", args, &reply)
	return nil
}
