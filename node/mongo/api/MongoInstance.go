package mongo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"progetto-sdcc/node/mongo/communication"
	"progetto-sdcc/registry/amazon"
	"progetto-sdcc/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
var LAST_ACC string = "lastAcc"

/*
Struttura che mantiene una connessione verso una specifica collezione MongoDB
*/
type MongoInstance struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
	CloudKeys  []string
}

/*
Apre la connessione con il database, inizializzando la collection utilizzata
*/
func (cli *MongoInstance) OpenConnection() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	cli.Client = client
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	cli.Database = client.Database(DB_NAME)
	cli.Collection = cli.Database.Collection(COLL_NAME)
	utils.PrintTs("Connected to MongoDB!")
}

/*
Chiude la connessione con il database
*/
func (cli *MongoInstance) CloseConnection() {
	err := cli.Client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	utils.PrintTs("Connection to MongoDB closed.")
}

/*
Ritorna una entry specificando la sua chiave
*/
func (cli *MongoInstance) GetEntry(key string) *MongoEntry {
	utils.PrintHeaderL3("Mongo Get, Searching for: " + key)
	if utils.StringInSlice(key, cli.CloudKeys) {
		utils.PrintTs("Entry on Cloud System. Downloading...\n")
		cli.downloadEntryFromS3(key)
		cli.MergeCollection(utils.CLOUD_EXPORT_FILE, utils.CLOUD_RECEIVE_PATH+key+utils.CSV)
		cli.CloudKeys = utils.RemoveElement(cli.CloudKeys, key)
		utils.ClearDir(utils.CLOUD_EXPORT_PATH)
		utils.ClearDir(utils.CLOUD_RECEIVE_PATH)
	}

	coll := cli.Collection
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{primitive.E{Key: ID, Value: key}}).Decode(&result)
	entry := MongoEntry{}

	if err != nil {
		utils.PrintTs("Get Error: " + err.Error())
		return nil
	}
	id := result[ID].(string)
	value := result[VALUE].(string)
	timest := result[TIME].(primitive.DateTime)
	lastaccess, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	entry.Key = id
	entry.Value = value
	entry.Timest = timest.Time()
	entry.LastAcc = lastaccess

	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: LAST_ACC, Value: lastaccess}}}}
	cli.Collection.UpdateOne(context.TODO(), entry, update)
	utils.PrintTs("Found: " + entry.Format())
	return &entry
}

/*
Legge una entry senza effettuare un accesso effettivo alla risorsa. Utile per identificare le entry raramente utilizzate
*/
func (cli *MongoInstance) ReadEntry(key string) *MongoEntry {
	coll := cli.Collection
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{primitive.E{Key: ID, Value: key}}).Decode(&result)
	if err != nil {
		utils.PrintTs("Read Error: " + err.Error())
		return nil
	}
	entry := MongoEntry{}
	id := result[ID].(string)
	value := result[VALUE].(string)
	timest := result[TIME].(primitive.DateTime)

	lastAcc := result[LAST_ACC].(primitive.DateTime)
	entry.Key = id
	entry.Value = value
	entry.Timest = timest.Time()
	entry.LastAcc = lastAcc.Time()
	utils.PrintTs("Read:" + entry.Format())
	return &entry
}

/*
Inserisce un'entry, specificando la chiave ed il suo valore.
Al momento del get viene calcolato il timestamp
*/
func (cli *MongoInstance) PutEntry(key string, value string) error {
	entry := fmt.Sprintf("{ %s , %s }", key, value)
	utils.PrintHeaderL3("Mongo Put, inserting " + entry)
	coll := cli.Collection
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	strVal := utils.FormatValue(value)
	doc := bson.D{primitive.E{Key: ID, Value: key}, primitive.E{Key: VALUE, Value: strVal},
		primitive.E{Key: TIME, Value: timestamp}, primitive.E{Key: LAST_ACC, Value: timestamp}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			utils.PrintTs("An entry for key " + key + " is already present on local storage")
			utils.PrintTs("Updating Entry Value")
			old := bson.D{primitive.E{Key: ID, Value: key}}
			timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
			update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: VALUE, Value: strVal},
				primitive.E{Key: TIME, Value: timestamp}, primitive.E{Key: LAST_ACC, Value: timestamp}}}}
			_, err := cli.Collection.UpdateOne(context.TODO(), old, update)
			if err != nil {
				utils.PrintTs(err.Error())
				return err
			}
			utils.PrintTs("Update: Entry for key " + key + ", updated into " + entry)
			return errors.New("Updated")

		} else {
			utils.PrintTs("Put Error: " + err.Error())
		}
		return err
	}
	utils.PrintTs("Entry " + entry + " succesfully inserted into local storage")
	return nil
}

/*
Aggiorna un'entry del database, specificando la chiave ed il nuovo valore da aggiungere.
Viene inoltre aggiornato il timestamp di quell'entry
*/
func (cli *MongoInstance) AppendValue(key string, arg1 string) error {
	utils.PrintHeaderL3("Mongo Append, adding argument " + arg1 + " to key " + key)
	old := bson.D{primitive.E{Key: ID, Value: key}}
	oldEntry := cli.GetEntry(key)
	if oldEntry == nil {
		utils.PrintTs("Append Error: No entry found with key " + key)
		return errors.New("NoKeyFound")
	}
	append := utils.AppendValue(oldEntry.Value, arg1)
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: VALUE, Value: append},
		primitive.E{Key: TIME, Value: timestamp}, primitive.E{Key: LAST_ACC, Value: timestamp}}}}
	_, err := cli.Collection.UpdateOne(context.TODO(), old, update)
	if err != nil {
		utils.PrintTs(err.Error())
		return err
	}
	utils.PrintTs("Append: inserted " + arg1 + " to key " + key)
	return nil
}

/*
Cancella un'entry dal database, specificandone la chiave
*/
func (cli *MongoInstance) DeleteEntry(key string) error {
	utils.PrintHeaderL3("Mongo Delete, removing entry with key " + key)
	coll := cli.Collection
	entry := bson.D{primitive.E{Key: ID, Value: key}}
	result, err := coll.DeleteOne(context.TODO(), entry)
	if err != nil {
		utils.PrintTs("Delete Error: " + err.Error())
		return err
	}

	if result.DeletedCount == 1 {
		utils.PrintTs("Deleted " + key)
		return nil
	}
	utils.PrintTs("Delete Error: No entry found with key " + key)
	return errors.New("EntryNotFound")
}

/*
Cancella un database e tutte le sue collezioni
*/
func (cli *MongoInstance) DropDatabase() {
	err := cli.Database.Drop(context.TODO())
	if err != nil {
		utils.PrintTs(err.Error())
		return
	}
	utils.PrintTs("Local storage cleaned succesfully")
}

/*
Inserisce un oggetto MongoEntry nel db.
Utilizzata durante l'aggiornamento delle entry del DB locale
*/
func (cli *MongoInstance) PutMongoEntry(entry MongoEntry) {
	coll := cli.Collection
	key := entry.Key
	value := entry.Value
	timestamp := entry.Timest
	lastaccess := entry.LastAcc

	doc := bson.D{primitive.E{Key: ID, Value: key}, primitive.E{Key: VALUE, Value: value},
		primitive.E{Key: TIME, Value: timestamp}, primitive.E{Key: LAST_ACC, Value: lastaccess}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		utils.PrintTs("PutMongoEntry Error: " + err.Error())
		return
	}
}

/*
Esporta una collezione, scrivendola su un file csv
*/
func (cli *MongoInstance) ExportCollection(filename string) {
	app := "mongoexport"
	arg1 := "--collection=" + COLL_NAME
	arg2 := "--db=" + DB_NAME
	arg3 := "--type=csv"
	arg4 := "--fields=_id,value,timest,lastAcc"
	arg5 := "--out=" + filename

	cmd := exec.Command(app, arg1, arg2, arg3, arg4, arg5)
	_, err := cmd.Output()
	if err != nil {
		utils.PrintTs(err.Error())
		return
	}
	utils.PrintTs("MongoExport successfully: " + filename)
}

/*
Esporta una entry specifica in formato CSV.
*/
func (cli *MongoInstance) ExportDocument(key string, filename string) {
	app := "mongoexport"
	arg1 := "--collection=" + COLL_NAME
	arg2 := "--db=" + DB_NAME
	arg3 := "--type=csv"
	arg4 := "--fields=_id,value,timest,lastAcc"
	arg5 := "--query={_id : '" + key + "'}"
	arg6 := "--out=" + filename

	cmd := exec.Command(app, arg1, arg2, arg3, arg4, arg5, arg6)
	_, err := cmd.Output()
	if err != nil {
		utils.PrintTs(err.Error())
		return
	}
}

/*
Carica una chiave sul bucket s3, rimuovendola dal database locale
*/
func (cli *MongoInstance) uploadToS3(key string) {
	utils.PrintHeaderL3("Uploading Entry to S3")

	filename := key + ".csv"
	utils.PrintTs("Exporting csv " + filename)
	cli.ExportDocument(key, utils.CLOUD_EXPORT_PATH+filename)
	sess := communication.CreateSession()
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(utils.CLOUD_EXPORT_PATH + filename)
	if err != nil {
		utils.PrintTs("Open Error: " + err.Error())
		return
	}

	// Carica il file su S3
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(utils.BUCKET_NAME),
		Key:    aws.String(filename),
		Body:   f,
	})
	if err != nil {
		utils.PrintTs(err.Error())
		return
	}
	utils.PrintTs("Entry succesfully uploaded to cloud storage")

	// Caricato il file da s3 lo rimuovo in locale, e salvo il fatto che è presente sul cloud
	utils.PrintTs("Removing entry from local storage")
	cli.CloudKeys = append(cli.CloudKeys, key)
	cli.DeleteEntry(key)
	utils.PrintTs("Migration to S3 completed")

}

/*
Ottiene la chiave specificata dal bucket S3, salvandola in un file locale
*/
func (cli *MongoInstance) downloadEntryFromS3(key string) {
	utils.PrintHeaderL3("Downlaoding Entry from S3")
	sess := amazon.CreateSession()
	filename := key + utils.CSV
	downloader := s3manager.NewDownloader(sess)

	// Crea il file in cui verrà scritto l'oggetto scaricato da S3
	f, err := os.Create(utils.CLOUD_RECEIVE_PATH + filename)
	if err != nil {
		utils.PrintTs(fmt.Sprintf("failed to create file %q, %v", filename, err))
		return
	}

	// Scrive il contenuto dell'oggetto S3 sul file
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(utils.BUCKET_NAME),
		Key:    aws.String(filename),
	})
	if err != nil {
		utils.PrintTs(fmt.Sprintf("failed to download file, %v", err))
		return
	}
	utils.PrintTs("Entry succesfully retrieved form cloud storage")
}

/*
Routine che ogni ora controlla tutte le entry per vedere se è possibile
effettuare una migrazione delle risorse verso il cloud S3
*/
func (cli *MongoInstance) CheckRarelyAccessed() {
	for {
		time.Sleep(utils.RARELY_ACCESSED_CHECK_INTERVAL)
		opts := options.Find().SetSort(bson.D{primitive.E{Key: ID, Value: 1}})
		cursor, _ := cli.Collection.Find(context.TODO(), bson.D{}, opts)
		var results []bson.M

		if err := cursor.All(context.TODO(), &results); err != nil {
			log.Fatal(err)
		}
		utils.PrintHeaderL2("Check Rarely Acccessed Entries")
		for _, result := range results {
			key := result[ID].(string)
			entry := cli.ReadEntry(key)
			if entry != nil {
				timeNow, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
				diff := timeNow.Sub(entry.LastAcc)
				utils.PrintTs("Key " + key + " non-accessed since " + diff.String())
				if diff >= utils.RARELY_ACCESSED_TIME {
					utils.PrintTs("Entry not accessed for a long time. Migrating on Cloud")
					cli.uploadToS3(entry.Key)
				}
			}
		}
	}
}

/*
Invocata dalla goroutine ListenUpdates quando un nodo sta inviando le informazioni nel proprio DB
Effettua l'export del DB locale, si unisce il CSV con quello ricevuto e si aggiorna il DB.
*/
func (cli *MongoInstance) MergeCollection(exportFile string, receivedFile string) {
	utils.PrintHeaderL3("Merging mongo local storage")
	cli.ExportCollection(exportFile)
	localExport, local_err := ParseCSV(exportFile)
	receivedUpdate, recvd_err := ParseCSV(receivedFile)
	if local_err != nil || recvd_err != nil {
		return
	}
	mergedEntries := MergeEntries(localExport, receivedUpdate)
	cli.Collection.Drop(context.TODO())
	for _, entry := range mergedEntries {
		cli.PutMongoEntry(entry)
	}
	cli.Collection.Find(context.TODO(), nil)
	utils.PrintTs("Collection merged succesfully")
}

/*
Invocato quando si riceve un update di riconciliazione. Si utilizza
last-write-wins per risolvere i conflitti tra le entry
*/
func (cli *MongoInstance) ReconciliateCollection(exportFile string, receivedFile string) {
	utils.PrintHeaderL3("Resolving conflicts on mongo local storage")

	cli.ExportCollection(exportFile) // Dump del database Locale
	localExport, local_err := ParseCSV(exportFile)
	receivedUpdate, recvd_err := ParseCSV(receivedFile)
	if local_err != nil || recvd_err != nil {
		return
	}
	reconEntries := ReconciliateEntries(localExport, receivedUpdate)
	cli.Collection.Drop(context.TODO())
	for _, entry := range reconEntries {
		cli.PutMongoEntry(entry)
	}
	cli.Collection.Find(context.TODO(), nil)
	utils.PrintTs("Local collection reconciliated correctly")
}

/*
Inizializza il sistema di storage locale aprendo la connessione a MongoDB e lanciando
i listener e le routine per la gestione degli updates.
*/
func InitLocalSystem() MongoInstance {
	utils.PrintTs("Starting Mongo Local System")
	client := MongoInstance{}
	client.OpenConnection()

	// Inizializza un database vuoto, per eliminare eventuale documenti residui del nodo.
	client.DropDatabase()

	utils.PrintTs("Mongo is Up & Running")
	return client
}
