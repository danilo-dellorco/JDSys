package communication

import (
	"fmt"
	"io"
	"net"
	"os"
	"progetto-sdcc/utils"
	"strconv"
	"strings"
)

// Dimensione del buffer per trasferire il file di aggiornamento
const BUFFERSIZE = 1024

/*
Goroutine in cui ogni nodo è in attesa di connessioni per ricevere l'export CSV del DB di altri nodi
*/
func StartReceiver(fileChannel chan string) {
	server, err := net.Listen("tcp", utils.UPDATES_PORT)
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
		fmt.Println("Node connected!")
		receiveFile(connection, fileChannel)
	}
}

/*
Apre la connessione verso un altro nodo per trasmettere un file
*/
func StartSender(filename string, address string) {
	connection, err := net.Dial("tcp", address+utils.UPDATES_PORT)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Println("Ready to send DB export...")
	sendFile(connection, filename)
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
func sendFile(connection net.Conn, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
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
