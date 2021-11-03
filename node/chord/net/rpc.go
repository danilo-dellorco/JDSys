package net

import (
	"fmt"
	"progetto-sdcc/utils"
)

// Strutture per il passaggio dei parametri per le RPC dell'applicazione
type GetDeleteArgs struct {
	key string
}

type PutUpdateArgs struct {
	key   string
	value string
}

/*
Effettua la RPC per la Get di una key.
 1) il lookup per trovare il nodo che hosta una risorsa
 2) RPC effettiva di GET verso quel nodo chord
*/
func (s *ChordNode) GetRPC(args *GetDeleteArgs, reply *[]string) error {
	addr, err := Lookup(utils.HashString(args.key), s.successor.ipaddr)
	// [TODO] rpc.call(GetFuncRPC,addr)
	fmt.Println(addr, err)

	return nil
}
