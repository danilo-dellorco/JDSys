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
	utils.PrintHeaderL2("Received Get RPC for key " + args.Key)
	utils.PrintTs("Checking value on local storage...")
	entry := n.MongoClient.GetEntry(args.Key)
	if entry != nil {
		*reply = fmt.Sprintf("Key: %s\nValue: %s", entry.Key, entry.Value)
		return nil
	}

	utils.PrintTs("Key not found on local storage.")
	// senza successore non possiamo propagare la richiesta, il nodo potrebbe essere da solo e la chiave non c'è realmente,
	// oppure il gestore della chiave è un altro, quindi il client è costretto a riprovare in attesa che si ricostruisca
	// l'anello per recuperare effettivamente il valore associato alla chiave
	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	if succ == "" {
		*reply = "Key not found."
		return nil
	}

	utils.PrintTs("Forwarding Get Request on DHT")
	addr, _ := chord.Lookup(utils.HashString(args.Key), succ+utils.CHORD_PORT)
	client, _ := utils.HttpConnect(utils.RemovePort(addr), utils.RPC_PORT)
	utils.PrintTs("Request sent to: " + utils.ParseAddrRPC(addr))
	client.Call("Node.GetImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per inserire un'entry nello storage.
 1) Lookup per trovare il nodo che deve hostare la risorsa
 2) RPC effettiva di PUT verso quel nodo chord
*/
func (n *Node) PutRPC(args Args, reply *string) error {
	utils.PrintHeaderL2("Received Put RPC for key " + args.Key)

	me := n.ChordClient.GetIpAddress()
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)
	client, _ := utils.HttpConnect(utils.RemovePort(addr), utils.RPC_PORT)
	utils.PrintTs("Checking Key Handling")
	utils.PrintTs("Request sent to: " + utils.ParseAddrRPC(addr))

	client.Call("Node.PutImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per aggiornare un'entry nello storage.
 1) Lookup per trovare il nodo che hosta la risorsa
 2) RPC effettiva di APPEND verso quel nodo chord
*/
func (n *Node) AppendRPC(args Args, reply *string) error {
	utils.PrintHeaderL2("Received Append RPC for key " + args.Key)

	me := n.ChordClient.GetIpAddress()
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)
	client, _ := utils.HttpConnect(utils.RemovePort(addr), utils.RPC_PORT)

	utils.PrintTs("Checking Key Handling")
	utils.PrintTs("Request sent to: " + utils.ParseAddrRPC(addr))
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
	utils.PrintHeaderL2("Received Delete RPC for key " + args.Key)

	me := n.ChordClient.GetIpAddress()
	handlerNode, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)
	args.Handler = utils.RemovePort(handlerNode)
	args.Deleted = false

	client, _ := utils.HttpConnect(utils.RemovePort(handlerNode), utils.RPC_PORT)
	utils.PrintTs("Checking Key Handling")
	fmt.Println("Delete request forwarded to handling node:", utils.ParseAddrRPC(handlerNode))
	client.Call("Node.DeleteHandling", args, &reply)
	return nil
}

/*
Effettua il get. Scrive in reply la stringa contenente l'entry richiesta. Se l'entry
non è stata trovata restituisce un messaggio di errore.
*/
func (n *Node) GetImpl(args Args, reply *string) error {
	utils.PrintHeaderL2("Received Get RPC for key " + args.Key)
	utils.PrintTs("I'm the handling node")
	fmt.Println(args.Key)
	entry := n.MongoClient.GetEntry(args.Key)
	if entry == nil {
		*reply = "Entry not found"
	} else {
		*reply = fmt.Sprintf("Key: %s\nValue: %s", entry.Key, entry.Value)
	}
	utils.PrintTs(*reply)
	utils.PrintTs("Finished. Replying to caller")
	return nil
}

/*
Effettua il PUT. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (n *Node) PutImpl(args Args, reply *string) error {
	utils.PrintHeaderL2("Received Put RPC for key " + args.Key)
	utils.PrintTs("I'm the handling node")
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
	utils.PrintTs(*reply)
	utils.PrintTs("Finished. Replying to caller")

	// TODO fare questo in una goroutine perchè intanto rispondo alla RPC poi replico..
	// Se non ho avuto errori, se è presente il successore inviamo l'entry per fargli gestire una replica.
	utils.PrintTs("Sending Replica to successor")
	if ok {
		succ := n.ChordClient.GetSuccessor().GetIpAddr()
		if succ == "" {
			utils.PrintTs("Node hasn't a successor yet, data will be replicated later")
			return nil
		}
		n.MongoClient.ExportDocument(args.Key, utils.UPDATES_EXPORT_FILE)
		SendReplicationMsg(n, succ, "replication")
		utils.PrintTs("Replica sent Correctly")
	}
	return nil
}

/*
Effettua l'APPEND. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (n *Node) AppendImpl(args *Args, reply *string) error {
	utils.PrintHeaderL2("Received Append RPC for key " + args.Key)
	utils.PrintTs("I'm the handling node")
	arg1 := args.Key
	arg2 := args.Value
	err := n.MongoClient.AppendValue(arg1, arg2)
	if err == nil {
		*reply = "Value correctly appended"
		utils.PrintTs(*reply)
		utils.PrintTs("Forwarding replica updates to successor")

		// TODO forse in goroutine come Get Impl
		// Se non ho avuto errori, se è presente il successore inviamo l'entry per fargli gestire una replica.
		succ := n.ChordClient.GetSuccessor().GetIpAddr()
		if succ == "" {
			utils.PrintTs("Node hasn't a successor yet, data will be replicated later")
			return nil
		}
		n.MongoClient.ExportDocument(args.Key, utils.UPDATES_EXPORT_FILE)
		SendReplicationMsg(n, succ, "replication")
		utils.PrintTs("Replica sent Correctly")
	} else {
		*reply = "Entry not found"
		utils.PrintTs(*reply)
	}
	utils.PrintTs("Finished. Replying to caller")
	return nil
}

/*
Effettua il delete della risorsa sul nodo che deve gestirla.
Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (n *Node) DeleteHandling(args *Args, reply *string) error {
	utils.PrintHeaderL2("Received Delete RPC for key " + args.Key)
	utils.PrintTs("I'm the handling node")
	utils.PrintTs("Deleting value on local storage")
	err := n.MongoClient.DeleteEntry(args.Key)
	if err == nil {
		args.Deleted = true
		*reply = "Entry successfully deleted"
	} else {
		// Entry non è presente nel DB del nodo gestore, quindi non esiste
		if err.Error() == "Entry Not Found" {
			*reply = "The key searched for delete not exist"
			return nil
		}
	}
	utils.PrintTs(*reply)

	// Se l'entry esiste ed è stata cancellata, procediamo inoltrando la richiesta al nodo successore
	// così da eliminare tutte le repliche nell'anello
	// Se non è presente, il nodo potrebbe essere da solo, o un eventuale successore ancora non identificato
	// verrà poi aggiornato successivamente tramite la riconciliazione
	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	if succ == "" {
		fmt.Println(reply)
		return nil
	}
	client, _ := utils.HttpConnect(succ, utils.RPC_PORT)
	utils.PrintTs("Delete request forwarded to replication node: " + succ + utils.RPC_PORT)
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
		utils.PrintTs("Delete Request returned to the handling node")
		if args.Deleted {
			fmt.Println("Entry correctly removed from every node!")
			*reply = "Entry succesfully deleted"
		} else {
			*reply = "Entry to delete not found"
		}
		utils.PrintTs(*reply)
		return nil
	}
	utils.PrintHeaderL2("Received Delete RPC for key " + args.Key)
	// Cancella la richiesta sul db locale
	utils.PrintTs("Deleting replicated value on local storage")
	n.MongoClient.DeleteEntry(args.Key)

	// Propaga la Delete al nodo successivo, la cancellazione sul nodo che gestisce la chiave
	// è già stata effettuata, per questo se i nodi successivi non hanno successore aspettiamo
	// la ricostruzione della DHT Chord finchè non viene completata la Delete!

	utils.PrintTs("Forwarding delete request")
retry:
	succ := n.ChordClient.GetSuccessor().GetIpAddr()
	if succ == "" {
		utils.PrintTs("Node hasn't a successor, wait for the reconstruction...")
		time.Sleep(2 * time.Second)
		goto retry
	}
	client, _ := utils.HttpConnect(succ, utils.RPC_PORT)
	utils.PrintTs("Delete request forwarded to replication node: " + succ + utils.RPC_PORT)
	client.Call("Node.DeleteReplicating", args, &reply)
	return nil
}

// TODO continuare da qui Print Refactoring
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
