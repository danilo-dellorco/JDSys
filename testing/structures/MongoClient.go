package structures

import (
	"context"
	"fmt"
	"log"
	"os/exec"

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
var UPDATE_CSV string = "update.csv"
var LOCAL_CSV string = "local.csv"

type MongoClient struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
}

func (cli *MongoClient) CloseConnection() {
	err := cli.Client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}

func (cli *MongoClient) OpenConnection() {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	cli.Client = client

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	// Inizializza il database e la collection, siamo gia connessi a mongo
	cli.Database = client.Database(DB_NAME)
	cli.Collection = cli.Database.Collection(COLL_NAME)
	fmt.Println("Connected to MongoDB!")
}

func (cli *MongoClient) GetEntry(key string) *MongoEntry {
	coll := cli.Collection
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{{ID, key}}).Decode(&result)
	if err != nil {
		fmt.Println("Get Error:", err)
		return nil
	}

	entry := MongoEntry{}
	id := result[ID].(string)
	value := result[VALUE].(string)
	timest := result[TIME].(primitive.DateTime)
	entry.Key = id
	entry.Value = value
	entry.Timest = timest.Time()
	fmt.Println("Get: found", entry)
	return &entry
}

func (cli *MongoClient) PutEntry(key string, value string) {
	coll := cli.Collection
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	doc := bson.D{{ID, key}, {VALUE, value}, {TIME, timestamp}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("Put Error:", err)
		return
	}
	fmt.Println("Put: Entry {"+key, value+"} inserita correttamente nel database")
}

func (cli *MongoClient) PutMongoEntry(entry MongoEntry) {
	coll := cli.Collection
	key := entry.Key
	value := entry.Value
	timestamp := entry.Timest

	doc := bson.D{{ID, key}, {VALUE, value}, {TIME, timestamp}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("PutMongoEntry Error:", err)
		return
	}
	fmt.Println("PutMongoEntry: MongoEntry {"+key, value, timestamp.String()+"} inserita correttamente nel database")

}

func (cli *MongoClient) UpdateEntry(key string, newValue string) {
	old := bson.D{{ID, key}}
	oldValue := cli.GetEntry(key).Value
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	update := bson.D{{"$set", bson.D{{VALUE, newValue}, {TIME, timestamp}}}}
	_, err := cli.Collection.UpdateOne(context.TODO(), old, update)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Update:", key+", changed value from", oldValue, "to", newValue)
}

func (cli *MongoClient) DeleteEntry(key string) {
	coll := cli.Collection
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

func (cli *MongoClient) DropDatabase() {
	err := cli.Database.Drop(context.TODO())
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("Drop: Database", cli.Database.Name(), "dropped successfully")
}

func (cli *MongoClient) ExportCollection(filename string) {
	app := "mongoexport"
	arg1 := "--collection=" + COLL_NAME
	arg2 := "--db=" + DB_NAME
	arg3 := "--type=csv"
	arg4 := "--fields=_id,value,timest"
	arg5 := "--out=" + filename

	cmd := exec.Command(app, arg1, arg2, arg3, arg4, arg5)
	fmt.Println(cmd)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(stdout))
}

func (cli *MongoClient) ImportCollection(coll string) {
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

func (cli *MongoClient) TestQueries() {
	cli.DropDatabase()
	cli.PutEntry("MyKey3", "MyValue")
	cli.GetEntry("MyKey3")
	cli.UpdateEntry("MyKey3", "NewValue")
	cli.GetEntry("MyKey3")
	//cli.deleteEntry("MyKey")
	//cli.dropDatabase()
}
