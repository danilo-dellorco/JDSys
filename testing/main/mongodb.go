package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"progetto-sdcc/testing/structures"
	"strconv"
	"strings"
	"time"
)

const BUFFERSIZE = 1024

func main() {
	client := structures.MongoClient{}
	client.OpenConnection()

	mode := os.Args[1]

	if mode == "c" {
		RunClient()
	} else {
		SendUpdate(client)
	}

	client.CloseConnection()
}

func ParseCSV(file string) []structures.MongoEntry {
	csvFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened CSV file")
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		fmt.Println(err)
	}

	var entryList []structures.MongoEntry
	i := 0
	for _, line := range csvLines {
		if i == 0 {
			i++
			continue
		}

		timeString := line[2]
		tVal, _ := time.Parse(time.RFC3339, timeString)
		entry := structures.MongoEntry{Key: line[0], Value: line[1], Timest: tVal, Conflict: false}
		entryList = append(entryList, entry)
	}
	return entryList
}

func mergeEntries(local []structures.MongoEntry, update []structures.MongoEntry) []structures.MongoEntry {
	var mergedEntries []structures.MongoEntry

	for i := 0; i < len(local); i++ {
		for j := 0; j < len(update); j++ {
			var latestEntry structures.MongoEntry
			if local[i].Key == update[j].Key {
				local[i].Conflict = true
				update[j].Conflict = true
				if local[i].Timest.After(update[j].Timest) {
					latestEntry = local[i]
				} else {
					latestEntry = update[j]
				}
				mergedEntries = append(mergedEntries, latestEntry)
			}
		}
		if !local[i].Conflict {
			mergedEntries = append(mergedEntries, local[i])
		}
	}
	for _, u := range update {
		if !u.Conflict {
			mergedEntries = append(mergedEntries, u)
		}
	}
	return mergedEntries
}

func ReceiveUpdate() {
	server, err := net.Listen("tcp", "localhost"+":4321")
	if err != nil {
		fmt.Println("Error listetning: ", err)
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Server started! Waiting for updates from other nodes ...")
}

func UpdateCollection(cli structures.MongoClient) {
	cli.ExportCollection(structures.LOCAL_CSV)
	RunClient() // contatta il nodo ed ottiene update.csv
	localList := ParseCSV(structures.LOCAL_CSV)
	updateList := ParseCSV(structures.UPDATE_CSV)
	mergedList := mergeEntries(localList, updateList)
	cli.Collection.Drop(context.TODO())
	for _, entry := range mergedList {
		cli.PutMongoEntry(entry)
	}
	cli.Collection.Find(context.TODO(), nil)
}

func SendUpdate(cli structures.MongoClient) {
	cli.ExportCollection(structures.UPDATE_CSV)
	sendUpdate(structures.UPDATE_CSV)

}

func sendUpdate(filename string) {
	server, err := net.Listen("tcp", "localhost:27001")
	if err != nil {
		fmt.Println("Error listetning: ", err)
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Server started! Waiting for connections...")
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		fmt.Println("Client connected")
		go sendFileToClient(connection, filename)
	}
}

func sendFileToClient(connection net.Conn, filename string) {
	fmt.Println("A client has connected!")
	defer connection.Close()
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
	fileName := fillString(fileInfo.Name(), 64)
	fmt.Println("Sending filename and filesize!")
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Start sending file!")
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	fmt.Println("File has been sent, closing connection!")
	return
}

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

func RunClient() {
	connection, err := net.Dial("tcp", "localhost:27001")
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Println("Connected to server, start receiving the file name and file size")
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	newFile, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, connection, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}
	fmt.Println("Received file completely!")
}
