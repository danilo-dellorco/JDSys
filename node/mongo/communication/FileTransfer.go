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
func StartReceiver(fileChannel chan string, mode string) {
	var port string
	switch mode {
	case "replication":
		port = utils.FILETR_TERMINATING_PORT
	case "reconciliation":
		port = utils.FILETR_RECONCILIATION_PORT
	}

	server, err := net.Listen("tcp", port)
	if err != nil {
		utils.PrintTs("Listening Error: " + err.Error())
		os.Exit(1)
	}
	defer server.Close()
	for {
		connection, err := server.Accept()
		if err != nil {
			utils.PrintTs("Error: " + err.Error())
			os.Exit(1)
		}
		if mode == "replication" {
			utils.PrintHeaderL3("A node wants to send his updates via TCP")
		} else {
			fmt.Println("mode:", mode)
			utils.PrintHeaderL3("A node wants to send a Reconciliation message via TCP")
		}
		receiveFile(connection, fileChannel, mode)
	}
}

/*
Apre la connessione verso un altro nodo per trasmettere un file
*/
func StartSender(filename string, address string, mode string) error {
	var addr string
	switch mode {
	case "replication":
		addr = address + utils.FILETR_TERMINATING_PORT
	case "reconciliation":
		addr = address + utils.FILETR_RECONCILIATION_PORT
	}
	connection, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	utils.PrintTs("Ready to send DB export")
	return sendFile(connection, filename)
}

/*
Utility per ricevere un file tramite la connessione
*/
func receiveFile(connection net.Conn, fileChannel chan string, mode string) {
	var receivedBytes int64
	var newFile *os.File
	var err error

	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	switch mode {
	case "replication":
		newFile, err = os.Create(utils.REPLICATION_RECEIVE_FILE)
	case "reconciliation":
		newFile, err = os.Create(utils.RECONCILIATION_RECEIVE_FILE)
	}

	if err != nil {
		panic(err)
	}
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
	utils.PrintTs("File received correctly")
	fileChannel <- "rcvd"
}

/*
Utility per inviare un file tramite la connessione
*/
func sendFile(connection net.Conn, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		utils.PrintTs(err.Error())
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		utils.PrintTs(err.Error())
		return err
	}

	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	connection.Write([]byte(fileSize))
	sendBuffer := make([]byte, BUFFERSIZE)
	utils.PrintTs("Start sending file via TCP")
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	utils.PrintTs("File sent correctly!")
	return nil
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
