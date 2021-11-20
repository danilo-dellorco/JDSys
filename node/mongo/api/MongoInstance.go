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

var DeletedKeys []string

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
	utils.PrintFormattedTimestamp()
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
		fmt.Println("Get Error:", err)
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
	utils.PrintFormattedTimestamp()
	utils.PrintTs("Found: " + entry.Format())
	return &entry
}

// TODO Danilo: continuare da qui per il Print Reafactoring
/*
Legge una entry senza effettuare un accesso effettivo alla risorsa. Utile per identificare le entry raramente utilizzate
*/
func (cli *MongoInstance) ReadEntry(key string) *MongoEntry {
	coll := cli.Collection
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{primitive.E{Key: ID, Value: key}}).Decode(&result)
	if err != nil {
		fmt.Println("Read Error:", err)
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
	fmt.Println("Read:", entry)
	return &entry
}

/*
Inserisce un'entry, specificando la chiave ed il suo valore.
Al momento del get viene calcolato il timestamp
*/
func (cli *MongoInstance) PutEntry(key string, value string) error {
	utils.PrintFormattedTimestamp()
	fmt.Printf("PUT | Inserting {%s,%s}\n", key, value)
	coll := cli.Collection
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	strVal := utils.FormatValue(value)
	doc := bson.D{primitive.E{Key: ID, Value: key}, primitive.E{Key: VALUE, Value: strVal},
		primitive.E{Key: TIME, Value: timestamp}, primitive.E{Key: LAST_ACC, Value: timestamp}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			fmt.Printf("Put: Entry %s già presente nello storage\n", key)
			fmt.Println("Updating Entry Value...")
			old := bson.D{primitive.E{Key: ID, Value: key}}
			timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
			update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: VALUE, Value: strVal},
				primitive.E{Key: TIME, Value: timestamp}, primitive.E{Key: LAST_ACC, Value: timestamp}}}}
			_, err := cli.Collection.UpdateOne(context.TODO(), old, update)
			if err != nil {
				fmt.Println(err)
				return err
			}
			utils.PrintFormattedTimestamp()
			fmt.Println("Update:", key+", changed value into", value)
			return errors.New("Updated")

		} else {
			fmt.Println("Put Error:", err)
		}
		return err
	}
	utils.PrintFormattedTimestamp()
	fmt.Println("Put: Entry {"+key, value+"} inserita correttamente nel database")
	return nil
}

/*
Aggiorna un'entry del database, specificando la chiave ed il nuovo valore assegnato.
Viene inoltre aggiornato il timestamp di quell'entry
*/
func (cli *MongoInstance) AppendValue(key string, arg1 string) error {
	utils.PrintFormattedTimestamp()
	fmt.Printf("Append | Appending %s to %s\n", arg1, key)
	old := bson.D{primitive.E{Key: ID, Value: key}}
	oldEntry := cli.GetEntry(key)
	if oldEntry == nil {
		fmt.Println("Append Error: No entry found with key", key)
		return errors.New("NoKeyFound")
	}
	append := utils.AppendValue(oldEntry.Value, arg1)
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: VALUE, Value: append},
		primitive.E{Key: TIME, Value: timestamp}, primitive.E{Key: LAST_ACC, Value: timestamp}}}}
	_, err := cli.Collection.UpdateOne(context.TODO(), old, update)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Append: inserted", arg1, "to key", key)
	return nil
}

/*
Cancella un'entry dal database, specificandone la chiave
*/
func (cli *MongoInstance) DeleteEntry(key string) error {
	utils.PrintFormattedTimestamp()
	fmt.Printf("Delete | Deleting %s\n", key)
	coll := cli.Collection
	entry := bson.D{primitive.E{Key: ID, Value: key}}
	result, err := coll.DeleteOne(context.TODO(), entry)
	if err != nil {
		fmt.Println("Delete Error:", err)
		return err
	}

	if result.DeletedCount == 1 {
		fmt.Println("Delete: Cancellata", key)
		DeletedKeys = append(DeletedKeys, key)
		return nil
	}
	fmt.Println("Delete: non è stata trovata nessuna entry con chiave", key)
	return errors.New("Entry Not Found")
}

/*
Cancella un database e tutte le sue collezioni
*/
func (cli *MongoInstance) DropDatabase() {
	err := cli.Database.Drop(context.TODO())
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("Drop: Database", cli.Database.Name(), "dropped successfully")
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
		fmt.Println("PutMongoEntry Error:", err)
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
	filename := key + ".csv"
	cli.ExportDocument(key, utils.CLOUD_EXPORT_PATH+filename)
	fmt.Println("Starting S3 Upload")
	sess := communication.CreateSession()
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(utils.CLOUD_EXPORT_PATH + filename)
	if err != nil {
		fmt.Printf("Open Error: ")
		fmt.Println(err)
		return
	}

	// Carica il file su S3
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(utils.BUCKET_NAME),
		Key:    aws.String(filename),
		Body:   f,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("file uploaded to, %s\n", result.Location)

	// Caricato il file da s3 lo rimuovo in locale, e salvo il fatto che è presente sul cloud
	cli.CloudKeys = append(cli.CloudKeys, key)
	cli.DeleteEntry(key)
}

/*
Ottiene la chiave specificata dal bucket S3, salvandola in un file locale
*/
func (cli *MongoInstance) downloadEntryFromS3(key string) {
	sess := amazon.CreateSession()
	filename := key + utils.CSV
	downloader := s3manager.NewDownloader(sess)

	// Crea il file in cui verrà scritto l'oggetto scaricato da S3
	f, err := os.Create(utils.CLOUD_RECEIVE_PATH + filename)
	if err != nil {
		fmt.Printf("failed to create file %q, %v", filename, err)
		return
	}

	// Scrive il contenuto dell'oggetto S3 sul file
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(utils.BUCKET_NAME),
		Key:    aws.String(filename),
	})
	if err != nil {
		fmt.Printf("failed to download file, %v", err)
		return
	}
	fmt.Printf("file downloaded, %d bytes\n", n)
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
		fmt.Println("\n===== Check Rarely Accessed Files =====")
		for _, result := range results {
			key := result[ID].(string)
			entry := cli.ReadEntry(key)
			if entry != nil {
				timeNow, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
				diff := timeNow.Sub(entry.LastAcc)
				fmt.Println("Key", key, "non-accessed since:", diff)
				if diff >= utils.RARELY_ACCESSED_TIME {
					fmt.Println("Elemento Non acceduto da tanto, Migrazione su cloud...")
					cli.uploadToS3(entry.Key)
				}
			}
		}
		fmt.Print("=======================\n\n")
	}
}

/*
Invocata dalla goroutine ListenUpdates quando un nodo sta inviando le informazioni nel proprio DB
Effettua l'export del DB locale, si unisce il CSV con quello ricevuto e si aggiorna il DB.
*/
func (cli *MongoInstance) MergeCollection(exportFile string, receivedFile string) {
	cli.ExportCollection(exportFile) // Dump del database Locale
	localExport := ParseCSV(exportFile)
	receivedUpdate := ParseCSV(receivedFile)
	mergedEntries := MergeEntries(localExport, receivedUpdate)
	cli.Collection.Drop(context.TODO())
	for _, entry := range mergedEntries {
		cli.PutMongoEntry(entry)
	}
	cli.Collection.Find(context.TODO(), nil)
	fmt.Println("Local DB ReceivedCorrectly")
}

/*
Invocato quando si riceve un update di riconciliazione. Si utilizza
last-write-wins per risolvere i conflitti tra le entry
*/
func (cli *MongoInstance) ReconciliateCollection(exportFile string, receivedFile string) {
	utils.PrintTs("Starting Reconciliation")

	cli.ExportCollection(exportFile) // Dump del database Locale
	localExport := ParseCSV(exportFile)
	receivedUpdate := ParseCSV(receivedFile)
	reconEntries := ReconciliateEntries(localExport, receivedUpdate)
	cli.Collection.Drop(context.TODO())
	for _, entry := range reconEntries {
		cli.PutMongoEntry(entry)
	}
	cli.Collection.Find(context.TODO(), nil)
	utils.PrintTs("Local DB ReceivedCorrectly")
}

/*
Inizializza il sistema di storage locale aprendo la connessione a MongoDB e lanciando
i listener e le routine per la gestione degli updates.
*/
func InitLocalSystem() MongoInstance {
	utils.PrintTs("Starting Mongo Local System")
	client := MongoInstance{}
	client.OpenConnection()

	utils.PrintTs("Mongo is Up & Running")
	return client
}
