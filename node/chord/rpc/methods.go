package chord

import (
	chord "progetto-sdcc/node/chord/net"
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

// Metodo 1 dell'interfaccia
func (s *chord.ChordNode) GetRPC(args *GetDeleteArgs, reply *[]string) error {
	addr, err := chord.Lookup(utils.HashString(args.key), s.successor)

	instances := checkActiveNodes()
	var list = make([]string, len(instances))
	for i := 0; i < len(instances); i++ {
		list[i] = instances[i].PrivateIP
	}
	*reply = list
	return nil
}
