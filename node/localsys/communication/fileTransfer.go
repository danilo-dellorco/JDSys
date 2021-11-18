package communication

import (
	"fmt"
	"io"
	"net"
	"os"
	"progetto-sdcc/node/localsys/structures"
	"progetto-sdcc/utils"
	"strconv"
	"strings"
)

// Dimensione del buffer per trasferire il file di aggiornamento
const BUFFERSIZE = 1024

/*
Goroutine in cui ogni nodo è in attesa di connessioni per ricevere l'export CSV del DB di altri nodi
*/
func StartReceiver(fileChannel chan string, mode string) {
	var port string
	switch mode {
	case "update":
		port = utils.FILETR_TERMINATING_PORT
	case "reconciliation":
		port = utils.FILETR_RECONCILIATION_PORT
	}

	server, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error listening: ", err)
		os.Exit(1)
	}
	defer server.Close()
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		fmt.Println("Node want to send is updates!")
		receiveFile(connection, fileChannel)
	}
}

/*
Apre la connessione verso un altro nodo per trasmettere un file
*/
func StartSender(cli structures.MongoClient, filename string, address string, mode string) {
	var addr string
	switch mode {
	case "update":
		addr = address + utils.FILETR_TERMINATING_PORT
	case "reconciliation":
		addr = address + utils.FILETR_RECONCILIATION_PORT
	}

	connection, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Println("Ready to send DB export...")
	sendFile(cli, connection, filename)
}

/*
Utility per ricevere un file tramite la connessione
*/
func receiveFile(connection net.Conn, fileChannel chan string) {
	fmt.Println("Start receiving the filesize...")
	var receivedBytes int64
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	newFile, err := os.Create(utils.UPDATES_RECEIVE_FILE)

	if err != nil {
		panic(err)
	}
	fmt.Println("Start receiving file...")
	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, connection, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}
	defer newFile.Close()
	fmt.Println("File received correctly!")
	fileChannel <- "rcvd"
}

/*
Utility per inviare un file tramite la connessione
*/
func sendFile(cli structures.MongoClient, connection net.Conn, filename string) {
retry:
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		if err.Error() == "open exported.csv: no such file or directory" {
			fmt.Println("DB export removed from previous merge, retry creation and send to successor...")
			cli.ExportCollection(filename)
			goto retry
		}
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}

	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fmt.Println("Start sending the filesize...")
	connection.Write([]byte(fileSize))
	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Start sending file...")
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	fmt.Println("File send correctly!")
}

/*
Riempie una stringa per raggiungere una lunghezza specificata
*/
func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}
