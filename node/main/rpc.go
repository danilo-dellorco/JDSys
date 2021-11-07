package main

import (
	"fmt"
	chord "progetto-sdcc/node/chord/net"
	"progetto-sdcc/utils"
)

/*
Pseudo-Interfaccia che verrà registrata dal server in modo tale che il client possa invocare i metodi tramite RPC
ciò che si registra realmente è un oggetto che prevede l'implementazione di quei metodi specifici
*/
type RPCservice struct {
	node chord.ChordNode
}

/*
Parametri per le operazioni di Get e Delete
*/
type Args1 struct {
	key string
}

/*
Parametri per le operazioni di Put e Update
*/
type Args2 struct {
	key   string
	value string
}

/*
Effettua la RPC per la Get di una key.
 1) Lookup per trovare il nodo che hosta una risorsa
 2) RPC effettiva di GET verso quel nodo chord
*/
func (s *RPCservice) GetRPC(args *Args1, reply *[]string) error {
	node := s.node
	addr, err := chord.Lookup(utils.HashString(args.key), node.GetSuccessor().GetIpAddr())
	// [TODO] rpc.call(GetFuncRPC,addr)
	fmt.Println(addr, err)

	return nil
}
