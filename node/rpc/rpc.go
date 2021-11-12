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

/*
Interfaccia registrata dal nodo in modo tale che il client possa invocare i metodi tramite RPC
Ciò che poi si registra realmente è un oggetto che ha l'implementazione dei precisi metodi offerti

I metodi di Get,Put,Delete,Update vengono invocati tramite RPC dai client
Ricevuta la richiesta, il nodo effettua il lookup per trovare chi mantiene la risorsa
--> Seconda RPC verso l'effettivo nodo che gestisce la chiave cercata!
*/
type RPCservice struct {
	Node chord.ChordNode
	Db   structures.MongoClient
}

/*
Struttura per l'RPC effettiva
*/
type ImplArgs struct {
	Key   string
	Value string
	ip    string
}

/*
Parametri per le operazioni di Get e Delete
*/
type Args1 struct {
	Key string
}

/*
Parametri per le operazioni di Put e Update
*/
type Args2 struct {
	Key   string
	Value string
}

/*
Effettua la RPC per la Get di una Key.
 1) Lookup per trovare il nodo che hosta una risorsa
 2) RPC effettiva di GET verso quel nodo chord
*/
func (s *RPCservice) GetRPC(args *Args1, reply *string) error {
	fmt.Println("GetRPC called!")

	me := s.Node.GetIpAddress()

	//porta 4567 per lookup di Chord
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)

	//porta 80 per RPC dell'applicazione
	//lookup ritorna IP+porta, quindi dobbiamo toglierla e inserire quella su cui fare RPC
	client, err := rpc.DialHTTP("tcp", addr[:len(addr)-5]+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Request send to:", addr[:len(addr)-5])
	client.Call("RPCservice.GetImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per inserire un'entry nello storage.
 1) Lookup per trovare il nodo che deve hostare la risorsa
 2) RPC effettiva di PUT verso quel nodo chord
*/
func (s *RPCservice) PutRPC(args *Args2, reply *string) error {
	fmt.Println("PutRPC Called!")

	me := s.Node.GetIpAddress()

	//porta 4567 per lookup di Chord
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)

	//porta 80 per RPC dell'applicazione
	//lookup ritorna IP+porta, quindi dobbiamo toglierla e inserire quella su cui fare RPC
	client, err := rpc.DialHTTP("tcp", addr[:len(addr)-5]+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Request send to:", addr[:len(addr)-5])
	client.Call("RPCservice.PutImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per aggiornare un'entry nello storage.
 1) Lookup per trovare il nodo che hosta la risorsa
 2) RPC effettiva di UPDATE verso quel nodo chord
*/
func (s *RPCservice) UpdateRPC(args *Args2, reply *string) error {
	fmt.Println("UpdateRPC Called!")
	me := s.Node.GetIpAddress()

	//porta 4567 per lookup di Chord
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)

	//porta 80 per RPC dell'applicazione
	//lookup ritorna IP+porta, quindi dobbiamo toglierla e inserire quella su cui fare RPC
	client, err := rpc.DialHTTP("tcp", addr[:len(addr)-5]+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Request send to:", addr[:len(addr)-5])
	client.Call("RPCservice.UpdateImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per eliminare un'entry nello storage.
 1) Lookup per trovare il nodo che hosta la risorsa
 2) RPC effettiva di DELETE verso quel nodo chord
*/
func (s *RPCservice) DeleteRPC(args *Args1, reply *string) error {
	fmt.Println("DeleteRPC called")
	me := s.Node.GetIpAddress()

	//porta 4567 per lookup di Chord
	addr, _ := chord.Lookup(utils.HashString(args.Key), me+utils.CHORD_PORT)

	//porta 80 per RPC dell'applicazione
	//lookup ritorna IP+porta, quindi dobbiamo toglierla e inserire quella su cui fare RPC
	client, err := rpc.DialHTTP("tcp", addr[:len(addr)-5]+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Request send to:", addr[:len(addr)-5])
	client.Call("RPCservice.DeleteImpl", args, &reply)
	return nil
}

/*
Effettua il get. Scrive in reply la stringa contenente l'entry richiesta. Se l'entry
non è stata trovata restituisce un messaggio di errore.
*/
func (s *RPCservice) GetImpl(args *Args1, reply *string) error {
	fmt.Println("Get request arrived")
	fmt.Println(args.Key)
	entry := s.Db.GetEntry(args.Key)
	fmt.Println(entry.Value)
	if entry.Value == "" {
		*reply = "Entry not found"
	} else {
		*reply = fmt.Sprintf("Key: %s\nValue: %s", entry.Key, entry.Value)
	}
	return nil
}

/*
Effettua il PUT. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (s *RPCservice) PutImpl(args *Args2, reply *string) error {
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
Effettua l'UPDATE. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (s *RPCservice) UpdateImpl(args *Args2, reply *string) error {
	fmt.Println("Update request arrived")
	arg1 := args.Key
	arg2 := args.Value
	fmt.Println("Arguments", arg1, arg2)
	err := s.Db.UpdateEntry(arg1, arg2)
	if err == nil {
		*reply = "Entry correctly updated"
	} else {
		*reply = "Entry not found"
	}
	return nil
}

/*
Effettua il DELETE. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (s *RPCservice) DeleteImpl(args *Args1, reply *string) error {
	fmt.Println("Delete request arrived")
	err := s.Db.DeleteEntry(args.Key)
	if err == nil {
		*reply = "Entry correctly deleted"
	} else {
		*reply = "Entry to delete not found"
	}
	return nil
}

/*
Metodo invocato dal Service Registry quando l'istanza EC2 viene schedulata per la terminazione
Effettua il trasferimento del proprio DB al nodo successore nella rete per garantire replicazione dei dati
*/

func (s *RPCservice) TerminateInstanceRPC(args *Args1, reply *string) error {
	addr := s.Node.GetSuccessor().GetIpAddr()
	fmt.Println("Instance Scheduled to Terminating...")
	mongo.SendUpdate(s.Db, addr)
	*reply = "Instance Terminating"
	return nil
}
