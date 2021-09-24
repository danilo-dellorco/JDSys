package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB_NAME string = "sdcc_storage_sys"
var COLL_NAME string = "sdcc_storage_local"
var ID string = "_id"
var VALUE string = "value"

type mongoEntry struct {
	_id   string
	value string
}

type mongoSearch struct {
	field string
	value string
}

type mongoClient struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func (db *mongoClient) closeConnection() {
	err := db.client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}

func (db *mongoClient) openConnection() {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	db.client = client

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	// Inizializza il database e la collection, siamo gia connessi a mongo
	db.database = client.Database(DB_NAME)
	db.collection = db.database.Collection(COLL_NAME)
	fmt.Println("Connected to MongoDB!")
}

func (db *mongoClient) getEntry(key string) {
	coll := db.collection
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{{ID, key}}).Decode(&result)
	if err != nil {
		fmt.Println("Get Error:", err)
		return
	}

	entry := mongoEntry{}
	id := result[ID].(string)
	value := result[VALUE].(string)
	entry._id = id
	entry.value = value
	fmt.Println("Get: found", entry)
}

func (db *mongoClient) putEntry(n mongoEntry) {
	coll := db.collection
	doc := bson.D{{ID, n._id}, {VALUE, n.value}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("Put Error:", err)
		return
	}
	fmt.Println("Put: Entry", n, "inserita correttamente nel DB")

}

func (db *mongoClient) deleteEntry(key string) {
	coll := db.collection
	entry := bson.D{{ID, key}}
	result, err := coll.DeleteOne(context.TODO(), entry)
	if err != nil {
		fmt.Println("Delete Error:", err)
		return
	}
	//TODO vedere return 1 o 0 per vedere se ha cancellato oppure no
	if result.DeletedCount == 1 {
		fmt.Println("Delete: Cancellata entry con chiave", key)
		return
	}
	fmt.Println("Delete: non è stata trovata nessuna entry con chiave", key)
}

func (db *mongoClient) dropDatabase() {
	err := db.database.Drop(context.TODO())
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("Drop: Database", db.database.Name(), "dropped successfully")
}

func main() {
	mg := mongoClient{}
	mg.openConnection()

	// Testing put, get and delete of an entry
	///*
	entry := mongoEntry{_id: "MyKey", value: "MyValue"}
	mg.putEntry(entry)
	mg.getEntry("MyKey")
	mg.deleteEntry("MyKey")
	mg.dropDatabase()
	//*/
	mg.closeConnection()
}
