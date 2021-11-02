package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	chord "progetto-sdcc/node/chord/net"
	mongo "progetto-sdcc/node/mongo/core"
	"time"
)

type EmptyArgs struct{}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Wrong usage: Specify registry private IP address")
		return
	}

	//start receiving heartbeats from LB
	go HealthyCheck()

	//setup flags
	addressPtr := flag.String("addr", "", "the port you will listen on for incomming messages")
	joinPtr := flag.String("join", "", "an address of a server in the Chord network to join to")
	flag.Parse()

	//get IP of the host used in the VPC
	*addressPtr = GetOutboundIP().String() + ":4567"
	me := new(chord.ChordNode)

	//wait to become healthy before join the Chord Network
	time.Sleep(time.Minute)

	//check active instances contacting the service registry
	//do it while there is at least one healthy instance
	result := JoinDHT(os.Args[1])
	for {
		if len(result) == 0 {
			result = JoinDHT(os.Args[1])
		} else {
			break
		}
	}
	fmt.Println(result)
	fmt.Println(len(result))

	//one active instance, me, so create a new ring
	if len(result) == 1 {
		me = chord.Create(*addressPtr)
	} else {
		//found active instances, join the ring contacting a random node excluse me
		*joinPtr = result[rand.Intn(len(result))]
		for {
			if *joinPtr == *addressPtr {
				*joinPtr = result[rand.Intn(len(result))]
			} else {
				break
			}
		}
		*joinPtr = *joinPtr + ":4567"
		me, _ = chord.Join(*addressPtr, *joinPtr)
	}
	fmt.Printf("My address is: %s.\n", *addressPtr)
	fmt.Printf("Join address is: %s.\n", *joinPtr)

	//[TODO] Vedere bene dove metterlo. inizializza il database locale e tutte le routine di aggiornamento.
	mongo.InitLocalSystem()

	// [TODO] Togliere, sono stampe di debug ma il nodo non riceve comandi da riga di comando ma tramite RPC
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

func home_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Homepage")
}

// On port 8888, the node receives heartbeats from LB, configured on the aws target group
// sulla porta 80 serviremo le rpc dell'app
func HealthyCheck() {
	http.HandleFunc("/", home_handler)
	http.ListenAndServe(":8888", nil)
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
