package main

import (
	"context"
	"fmt"
	"log"

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
	_id    string
	value  string
	timest primitive.DateTime
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
	entry.timest = timest
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

func (cli *mongoClient) testQueries() {
	cli.putEntry("MyKey", "MyValue")
	cli.getEntry("MyKey")
	cli.updateEntry("MyKey", "NewValue")
	cli.getEntry("MyKey")
	cli.deleteEntry("MyKey")
	cli.dropDatabase()
}

func main() {
	client := mongoClient{}
	client.openConnection()
	client.testQueries()
	client.closeConnection()
}
