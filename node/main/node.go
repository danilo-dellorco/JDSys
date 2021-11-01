package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	chord "progetto-sdcc/node"
)

type EmptyArgs struct{}

func HttpConnect(serverAddress string) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", serverAddress+":1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	return client, err
}

func JoinDHT(serverAddress string) []string {
	args := EmptyArgs{}
	var reply []string

	client, _ := HttpConnect(serverAddress)
	err := client.Call("DHThandler.JoinRing", args, &reply)
	if err != nil {
		log.Fatal("RPC error: ", err)
	}
	return reply
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func main() {
	//if len(os.Args) < 2 {
	//	fmt.Println("Wrong usage: Specify registry IP address")
	//	return
	//}

	//setup flags
	addressPtr := flag.String("addr", "", "the port you will listen on for incomming messages")
	joinPtr := flag.String("join", "", "an address of a server in the Chord network to join to")
	flag.Parse()

	//get IP of the host used in the VPC
	//*addressPtr = GetOutboundIP().String() + ":4567"
	*addressPtr = "mio IP"
	*joinPtr = "IP nodo tramite cui entrare nella rete chord"
	me := new(chord.ChordNode)

	//check active instances contacting the service registry
	//result := JoinDHT(os.Args[1])
	//result := JoinDHT("3.95.38.29")
	//fmt.Println(result)

	//one active instance, me, so create a new ring
	//if len(result) == 1 {
	me = chord.Create(*addressPtr)
	//} else {
	//found active instances, join the ring contacting a random node
	//*joinPtr = result[rand.Intn(len(result))] + ":4567"
	//fmt.Println(*joinPtr)
	//me, _ = chord.Join(*addressPtr, *joinPtr)
	//}
	fmt.Printf("My address is: %s.\n", *addressPtr)
	fmt.Printf("Join address is: %s.\n", *joinPtr)

	//block until receive input
Loop:
	for {
		var cmd string
		_, err := fmt.Scan(&cmd)
		switch {
		case cmd == "print":
			//print out successor and predecessor
			fmt.Printf("%s", me.String())
		case cmd == "fingers":
			//print out finger table
			fmt.Printf("%s", me.ShowFingers())
		case cmd == "succ":
			//print out successor list
			fmt.Printf("%s", me.ShowSucc())
		case err == io.EOF:
			break Loop
		}

	}
	me.Finalize()
}
