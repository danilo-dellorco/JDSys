package structures

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"progetto-sdcc/registry/services"
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
type MongoClient struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
	CloudKeys  []string
}

/*
Apre la connessione con il database, inizializzando la collection utilizzata
*/
func (cli *MongoClient) OpenConnection() {
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
	fmt.Println("Connected to MongoDB!")
}

/*
Chiude la connessione con il database
*/
func (cli *MongoClient) CloseConnection() {
	err := cli.Client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}

/*
Ritorna una entry specificando la sua chiave
*/
func (cli *MongoClient) GetEntry(key string) *MongoEntry {
	if utils.StringInSlice(key, cli.CloudKeys) {
		fmt.Printf("Entry %s presente nel cloud. Downloading...\n", key)
		cli.downloadEntryFromS3(key)
		cli.UpdateCollection(utils.CLOUD_EXPORT_FILE, utils.CLOUD_RECEIVE_PATH+key+utils.CSV)
		cli.CloudKeys = utils.RemoveElement(cli.CloudKeys, key)
		utils.ClearDir(utils.CLOUD_EXPORT_PATH)
		utils.ClearDir(utils.CLOUD_RECEIVE_PATH)
	}

	coll := cli.Collection
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{{ID, key}}).Decode(&result)
	fmt.Println(result)
	if err != nil {
		fmt.Println("Get Error:", err)
		return nil
	}
	entry := MongoEntry{}
	id := result[ID].(string)
	value := result[VALUE].(string)
	timest := result[TIME].(primitive.DateTime)
	lastaccess, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	entry.Key = id
	entry.Value = value
	entry.Timest = timest.Time()
	entry.LastAcc = lastaccess

	update := bson.D{{"$set", bson.D{{LAST_ACC, lastaccess}}}}
	_, err = cli.Collection.UpdateOne(context.TODO(), entry, update)
	fmt.Println("Get: found", entry)
	return &entry
}

/*
Legge una entry senza effettuare un accesso effettivo alla risorsa. Utile per
la identificare le entry raramente utilizzate
*/
func (cli *MongoClient) ReadEntry(key string) *MongoEntry {
	coll := cli.Collection
	var result bson.M
	err := coll.FindOne(context.TODO(), bson.D{{ID, key}}).Decode(&result)
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
func (cli *MongoClient) PutEntry(key string, value string) {
	coll := cli.Collection
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	doc := bson.D{{ID, key}, {VALUE, value}, {TIME, timestamp}, {LAST_ACC, timestamp}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		if strings.Contains(err.Error(), "E11000") {
			fmt.Printf("Entry %s già presente nello storage\n", key)
			// [TODO] Overwrite oppure no delle Entry già presenti?
			// In futuro si, ora per debug meglio di no
		} else {
			fmt.Println("Put Error:", err)
		}
		return
	}
	fmt.Println("Put: Entry {" + key + "} inserita correttamente nel database")
}

/*
Aggiorna un'entry del database, specificando la chiave ed il nuovo valore assegnato.
Viene inoltre aggiornato il timestamp di quell'entry
*/
func (cli *MongoClient) UpdateEntry(key string, newValue string) {
	old := bson.D{{ID, key}}
	oldValue := cli.GetEntry(key).Value
	timestamp, _ := ntp.Time("0.beevik-ntp.pool.ntp.org")
	update := bson.D{{"$set", bson.D{{VALUE, newValue}, {TIME, timestamp}, {LAST_ACC, timestamp}}}}
	_, err := cli.Collection.UpdateOne(context.TODO(), old, update)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Update:", key+", changed value from", oldValue, "to", newValue)
}

/*
Cancella un'entry dal database, specificandone la chiave
*/
func (cli *MongoClient) DeleteEntry(key string) {
	coll := cli.Collection
	entry := bson.D{{ID, key}}
	result, err := coll.DeleteOne(context.TODO(), entry)
	if err != nil {
		fmt.Println("Delete Error:", err)
		return
	}

	if result.DeletedCount == 1 {
		fmt.Println("Delete: Cancellata", key)
		return
	}
	fmt.Println("Delete: non è stata trovata nessuna entry con chiave", key)
}

/*
Cancella un database e tutte le sue collezioni
*/
func (cli *MongoClient) DropDatabase() {
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
func (cli *MongoClient) PutMongoEntry(entry MongoEntry) {
	coll := cli.Collection
	key := entry.Key
	value := entry.Value
	timestamp := entry.Timest
	lastaccess := entry.LastAcc

	doc := bson.D{{ID, key}, {VALUE, value}, {TIME, timestamp}, {LAST_ACC, lastaccess}}
	_, err := coll.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("PutMongoEntry Error:", err)
		return
	}
}

/*
Esporta una collezione, scrivendola su un file csv
*/
func (cli *MongoClient) ExportCollection(filename string) {
	app := "mongoexport"
	arg1 := "--collection=" + COLL_NAME
	arg2 := "--db=" + DB_NAME
	arg3 := "--type=csv"
	arg4 := "--fields=_id,value,timest,lastAcc"
	arg5 := "--out=" + filename

	cmd := exec.Command(app, arg1, arg2, arg3, arg4, arg5)
	_, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

/*
Esporta una entry specifica in formato CSV.
*/
func (cli *MongoClient) ExportDocument(key string, filename string) {
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
		fmt.Println(err.Error())
		return
	}
}

/*
Carica una chiave sul bucket s3, rimuovendola dal database locale
*/
func (cli *MongoClient) uploadToS3(key string) {
	filename := key + ".csv"
	cli.ExportDocument(key, utils.CLOUD_EXPORT_PATH+filename)
	fmt.Println("Starting S3 Upload")
	sess := services.CreateSession()
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
func (cli *MongoClient) downloadEntryFromS3(key string) {
	sess := services.CreateSession()
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
Goroutine in attesa di ricevere aggiornamenti remoti. Ogni volta che si riceve un CSV da un
nodo remoto viene aggiornato il database locale.
*/
func (cli *MongoClient) UpdateCollection(exportFile string, receivedFile string) {
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
Routine che periodicamente controlla tutte le entry per vedere se è possibile
effettuare una migrazione delle risorse verso il cloud S3
*/
func (cli *MongoClient) CheckRarelyAccessed() {
	opts := options.Find().SetSort(bson.D{{"_id", 1}})
	cursor, err := cli.Collection.Find(context.TODO(), bson.D{}, opts)
	var results []bson.M

	if err = cursor.All(context.TODO(), &results); err != nil {
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
