package utils

import "time"

// MongoDB Settings
var CSV string = ".csv"
var CLOUD_EXPORT_PATH string = "../mongo/communication/cloud/export/"
var CLOUD_RECEIVE_PATH string = "../mongo/communication/cloud/receive/"
var UPDATES_EXPORT_PATH string = "../mongo/communication/updates/export/"
var UPDATES_RECEIVE_PATH string = "../mongo/communication/updates/receive/"
var CLOUD_EXPORT_FILE string = CLOUD_EXPORT_PATH + "exported.csv"
var UPDATES_EXPORT_FILE string = UPDATES_EXPORT_PATH + "exported.csv"
var UPDATES_RECEIVE_FILE string = UPDATES_RECEIVE_PATH + "received.csv"

// AWS SDK Settings
var ELB_ARN string = "arn:aws:elasticloadbalancing:us-east-1:786781699181:loadbalancer/net/sdcc-lb/505f5d098d3c2bc3"
var AWS_CRED_PATH string = "/home/ec2-user/.aws/credentials"
var AUTOSCALING_NAME string = "sdcc-autoscaling"
var BUCKET_NAME string = "sdcc-cloud-resources"
var LB_DNS_NAME string = "sdcc-lb-505f5d098d3c2bc3.elb.us-east-1.amazonaws.com"
var REGISTRY_IP string = "10.0.0.64"

// Time Settings
// TODO impostare questi parametri a valori reali
var RARELY_ACCESSED_TIME time.Duration = 10 * time.Minute           // Dopo quanto tempo un'entry viene migrata sul cloud
var RARELY_ACCESSED_CHECK_INTERVAL time.Duration = 30 * time.Minute //Ogni quanto controlliamo entry vecchie
var NODE_HEALTHY_TIME time.Duration = 30 * time.Second              // Tempo di attesa di un nodo prima che diventi healthy
var NODE_SUCC_TIME time.Duration = 2 * time.Minute                  // Tempo di attesa di un nodo per essere sicuri che abbia il successore allo startup
var SEND_UPDATES_TIME time.Duration = time.Minute                   // Ogni quanto effettuare l'invio del backup del DB al nodo successore
var CHECK_TERMINATING_INTERVAL time.Duration = 30 * time.Second     // Ogni quanto effettuare il controllo sulle istanze in terminazione
var START_CONSISTENCY_INTERVAL time.Duration = 20 * time.Minute     // Ogni quanto avviare il processo di scambio di aggiornamenti tra i nodi per la consistenza finale
var ACTIVITY_CACHE_FLUSH_INTERVAL time.Duration = 40 * time.Minute  // Ogni quanto flushare la cache sulle istanze in terminazione
var CHORD_FIX_INTERVAL time.Duration = 10 * time.Second             // Ogni quanto un nodo contatta i suoi vicini per aggiornare le Finger Table
var RR1_TIMEOUT time.Duration = 100 * time.Millisecond              // Tempo dopo il quale si considera perso un messaggio client-server
var RR1_RETRIES = 5                                                 // Numero di ritrasmissioni RR1
var TEST_STEADY_TIME = 5 * time.Second                              // Tempo per inizializzare il workload nei test

// Port Settings
var HEARTBEAT_PORT string = ":8888"             // Porta su cui il nodo ascolta i segnali da load balancer e registry
var FILETR_TERMINATING_PORT string = ":7777"    // Porta su cui il nodo ascolta l'update mongo da altri nodi
var FILETR_RECONCILIATION_PORT string = ":6666" // Porta su cui il nodo ascolta l'update mongo da altri nodi
var FILETR_REPLICATION_PORT string = ":9999"    // Porta su cui il nodo riceve repliche di record da altri nodi
var RPC_PORT string = ":80"                     // Porta su cui il nodo ascolta le chiamate RPC
var REGISTRY_PORT string = ":4444"              // Porta tramite cui il nodo instaura una connessione con il Service Registry
var CHORD_PORT string = ":3333"                 // Porta tramite cui il nodo riceve ed invia i messaggi necessari ad aggiornare la DHT Chord
