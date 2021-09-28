package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/beevik/ntp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB_NAME string = "sdcc_storage_sys"
var COLL_NAME string = "sdcc_storage_local"
var ID string = "_id"
var VALUE string = "value"
var TIME string = "timest"

type mongoEntry struct {
	_id      string
	value    string
	timest   time.Time
	analyzed bool // rende piu efficiente il merge delle entry
}

func (me *mongoEntry) print() {
	fmt.Print("{" + me._id + ", " + me.value + ", " + me.timest.String() + "}")
	fmt.Printf(" %t\n", me.analyzed)
}

type mongoClient struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func (cli *mongoClient) closeConnection() {
	err := cli.client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}

func (cli *mongoClient) openConnection() {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	cli.client = client

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	// Inizializza il database e la collection, siamo gia connessi a mongo
	cli.database = client.Database(DB_NAME)
	cli.collection = cli.database.Collection(COLL_NAME)
	fmt.Println("Connected to MongoDB!")
}

func (cli *mongoClient) getEntry(key string) *mongoEntry {
	coll := cli.collection
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{{ID, key}}).Decode(&result)
	if err != nil {
		fmt.Println("Get Error:", err)
		return nil
	}

	entry := mongoEntry{}
	id := result[ID].(string)
	value := result[VALUE].(string)
	timest := result[TIME].(primitive.DateTime)
	entry._id = id
	entry.value = value
	entry.timest = timest.Time()
	fmt.Println("Get: found", entry)
	return &entry
}

func (cli *mongoClient) putEntry(key string, value string) {
	coll := cli.collection
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	doc := bson.D{{ID, key}, {VALUE, value}, {TIME, timestamp}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("Put Error:", err)
		return
	}
	fmt.Println("Put: Entry {"+key, value+"} inserita correttamente nel database")
}

func (cli *mongoClient) updateEntry(key string, newValue string) {
	old := bson.D{{ID, key}}
	oldValue := cli.getEntry(key).value
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	update := bson.D{{"$set", bson.D{{VALUE, newValue}, {TIME, timestamp}}}}
	_, err := cli.collection.UpdateOne(context.TODO(), old, update)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Update:", key+", changed value from", oldValue, "to", newValue)
}

func (cli *mongoClient) deleteEntry(key string) {
	coll := cli.collection
	entry := bson.D{{ID, key}}
	result, err := coll.DeleteOne(context.TODO(), entry)
	if err != nil {
		fmt.Println("Delete Error:", err)
		return
	}
	//TODO vedere return 1 o 0 per vedere se ha cancellato oppure no
	if result.DeletedCount == 1 {
		fmt.Println("Delete: Cancellata", key)
		return
	}
	fmt.Println("Delete: non è stata trovata nessuna entry con chiave", key)
}

func (cli *mongoClient) dropDatabase() {
	err := cli.database.Drop(context.TODO())
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("Drop: Database", cli.database.Name(), "dropped successfully")
}

func (cli *mongoClient) exportCollection(coll string) {
	app := "mongoexport"
	arg1 := "--collection=" + coll
	arg2 := "--db=sdcc_storage_sys"
	arg3 := "--type=csv"
	arg4 := "--fields=_id,value,timest"
	arg5 := "--out=export.csv"

	cmd := exec.Command(app, arg1, arg2, arg3, arg4, arg5)
	fmt.Println(cmd)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(stdout))
}

func (cli *mongoClient) importCollection(coll string) {
	app := "mongoimport"
	arg1 := "--db=sdcc_storage_sys"
	arg2 := "--collection=" + coll
	arg3 := "--file=export.json"

	cmd := exec.Command(app, arg1, arg2, arg3)
	fmt.Println(cmd)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(stdout))
}

func (cli *mongoClient) testQueries() {
	cli.dropDatabase()
	cli.putEntry("MyKey3", "MyValue")
	cli.getEntry("MyKey3")
	cli.updateEntry("MyKey3", "NewValue")
	cli.getEntry("MyKey3")
	//cli.deleteEntry("MyKey")
	//cli.dropDatabase()
}

func main() {
	client := mongoClient{}
	client.openConnection()
	//client.testQueries()
	//client.exportCollection(COLL_NAME)
	localList := ParseCSV("local.csv")
	updateList := ParseCSV("update.csv")
	mergeEntries(localList, updateList)
	// client.dropDatabase()
	// client.testMerge()
	client.closeConnection()
}

func ParseCSV(file string) []mongoEntry {
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

	var entryList []mongoEntry
	i := 0
	for _, line := range csvLines {
		if i == 0 {
			i++
			continue
		}

		timeString := line[2]
		tVal, _ := time.Parse(time.RFC3339, timeString)
		entry := mongoEntry{_id: line[0], value: line[1], timest: tVal, analyzed: false}
		entryList = append(entryList, entry)
	}
	return entryList
}

func mergeEntries(local []mongoEntry, update []mongoEntry) {
	var mergedEntries []mongoEntry

	for i := 0; i < len(local); i++ {
		fmt.Printf("LOCAL: ")
		local[i].print()
		for j := 0; j < len(update); j++ {
			fmt.Printf("UPDATE: ")
			update[j].print()
			var latestEntry mongoEntry
			if local[i]._id == update[j]._id {
				local[i].analyzed = true
				update[j].analyzed = true
				fmt.Println("Conflitto trovato!")
				fmt.Println("local: ", local[i]._id, local[i].value, local[i].timest.String())
				fmt.Println("update: ", update[j]._id, update[j].value, update[j].timest.String())
				if local[i].timest.After(latestEntry.timest) {
					latestEntry = local[i]
				} else {
					latestEntry = update[j]
				}
				fmt.Println("latest: ", latestEntry._id, latestEntry.value, latestEntry.timest.String())

				mergedEntries = append(mergedEntries, latestEntry)
			}
		}
		if !local[i].analyzed {
			mergedEntries = append(mergedEntries, local[i])
		}
	}
	for _, u := range update {
		if !u.analyzed {
			mergedEntries = append(mergedEntries, u)
		}
	}

	fmt.Println("Merged Entries")
	for _, entry := range mergedEntries {
		entry.print()
	}
}
