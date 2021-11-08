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
Pseudo-Interfaccia che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici
*/
type RPCservice struct {
	Node chord.ChordNode
	Db   structures.MongoClient
}

/*
Struttura per l'RPC effettiva dopo il
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
	fmt.Println("GetRPC called")

	// [TODO] rimettere il lookup. addr:=localhost solo per testing in locale
	//node := s.node
	//addr, err := Lookup(utils.HashString(args.Key), node.GetIpAddress())
	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	fmt.Println("Before Impl Call")
	err = client.Call("RPCservice.GetImpl", args, &reply)
	return nil
}

/*
Effettua la RPC per inserire un'entry nello storage.
 1) Lookup per trovare il nodo che deve hostare la risorsa
 2) RPC effettiva di PUT verso quel nodo chord
*/
func (s *RPCservice) PutRPC(args *Args2, reply *string) error {
	fmt.Println("PutRPC Called!")

	// [TODO] rimettere il lookup. addr:=localhost solo per testing in locale
	//node := s.node
	//addr, err := Lookup(utils.HashString(args.Key), node.GetIpAddress())
	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	err = client.Call("RPCservice.PutImpl", args, &reply)
	fmt.Println("int:", *reply)
	return nil
}

/*
Effettua la RPC per aggiornare un'entry nello storage.
 1) Lookup per trovare il nodo che hosta la risorsa
 2) RPC effettiva di UPDATE verso quel nodo chord
*/
func (s *RPCservice) UpdateRPC(args *Args2, reply *string) error {
	fmt.Println("UpdateRPC Called!")

	// [TODO] rimettere il lookup. addr:=localhost solo per testing in locale
	//node := s.node
	//addr, err := Lookup(utils.HashString(args.Key), node.GetIpAddress())
	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	err = client.Call("RPCservice.UpdateImpl", args, &reply)
	fmt.Println("int:", *reply)
	return nil
}

/*
Effettua la RPC per eliminare un'entry nello storage.
 1) Lookup per trovare il nodo che hosta la risorsa
 2) RPC effettiva di DELETE verso quel nodo chord
*/
func (s *RPCservice) DeleteRPC(args *Args1, reply *string) error {
	fmt.Println("DeleteRPC called")

	// [TODO] rimettere il lookup. addr:=localhost solo per testing in locale
	//node := s.node
	//addr, err := Lookup(utils.HashString(args.Key), node.GetIpAddress())
	addr := "localhost"

	client, err := rpc.DialHTTP("tcp", addr+utils.RPC_PORT)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	err = client.Call("RPCservice.DeleteImpl", args, &reply)
	return nil
}

/*
Effettua il get. Scrive in reply la stringa contenente l'entry richiesta. Se l'entry
non è stata trovata restituisce un messaggio di errore.
*/
func (s *RPCservice) GetImpl(args *Args1, reply *string) error {
	fmt.Println("GetImpl")
	fmt.Println(args.Key)
	entry := s.Db.GetEntry(args.Key)
	fmt.Println(entry.Value)
	if entry.Value == "" {
		*reply = "Entry not Found"
	} else {
		*reply = fmt.Sprintf("Key: %s\nValue: %s", entry.Key, entry.Value)
	}
	return nil
}

/*
Effettua il PUT. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (s *RPCservice) PutImpl(args *Args2, reply *string) error {
	fmt.Println("Called PutImpl")
	arg1 := args.Key
	arg2 := args.Value
	err := s.Db.PutEntry(arg1, arg2)
	if err == nil {
		*reply = "Entry inserita correttamente nel DB"
	} else {
		*reply = "Entry già presente nel DB"
		fmt.Println(*reply)
	}
	return nil
}

/*
Effettua l'UPDATE. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (s *RPCservice) UpdateImpl(args *Args2, reply *string) error {
	fmt.Println("Called UpdateImpl")
	arg1 := args.Key
	arg2 := args.Value
	fmt.Println("Arguments", arg1, arg2)
	err := s.Db.UpdateEntry(arg1, arg2)
	if err == nil {
		*reply = "Entry aggiornata correttamente"
	} else {
		*reply = "Entry not Found"
	}
	return nil
}

/*
Effettua il DELETE. Ritorna 0 se l'operazione è avvenuta con successo, altrimenti l'errore specifico
*/
func (s *RPCservice) DeleteImpl(args *Args1, reply *string) error {
	err := s.Db.DeleteEntry(args.Key)
	if err == nil {
		*reply = "Entry Deleted Succesfully"
	} else {
		*reply = "Entry to Delete not Found"
	}
	return nil
}

func (s *RPCservice) TerminateInstanceRPC(args *Args1, reply *string) error {
	addr := s.Node.GetSuccessor().GetIpAddr()
	fmt.Println("Instance Scheduled to Terminating...")
	mongo.SendUpdate(s.Db, addr)
	*reply = "Instance Terminating"
	return nil
}
