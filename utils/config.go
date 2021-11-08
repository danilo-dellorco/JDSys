package utils

import "time"

// MongoDB Settings
var CSV string = ".csv"
var CLOUD_EXPORT_PATH string = "../localsys/communication/cloud/export/"
var CLOUD_EXPORT_FILE string = CLOUD_EXPORT_PATH + "exported.csv"
var CLOUD_RECEIVE_PATH string = "../localsys/communication/cloud/receive/"
var UPDATES_EXPORT_PATH string = "../localsys/communication/updates/export/"
var UPDATES_RECEIVE_PATH string = "../localsys/communication/updates/receive/"
var UPDATES_EXPORT_FILE string = UPDATES_EXPORT_PATH + "exported.csv"
var UPDATES_RECEIVE_FILE string = UPDATES_RECEIVE_PATH + "received.csv"

// AWS SDK Settings
var ELB_ARN_D string = "arn:aws:elasticloadbalancing:us-east-1:427788101608:loadbalancer/net/NetworkLB/8d7f674bf6bc6f73"
var ELB_ARN_J string = "arn:aws:elasticloadbalancing:us-east-1:786781699181:loadbalancer/net/sdcc-lb/505f5d098d3c2bc3"
var AWS_CRED_PATH string = "/home/ec2-user/.aws/credentials"
var AWS_CRED_PATH_D string = "/home/danilo/.aws/credentials"
var AWS_CRED_PATH_J string = "/home/jacopo/.aws/credentials"
var AUTOSCALING_NAME_D string = "sdcc-autoscaling"
var AUTOSCALING_NAME_J string = "sdcc-autoscaling"
var BUCKET_NAME string = "sdcc-cloud-keys"

// Time Settings
var RARELY_ACCESSED_TIME time.Duration = 10                         // Dopo quanto tempo (ms) un'entry viene migrata sul cloud
var NODE_HEALTHY_TIME time.Duration = 20 * time.Second              // Tempo di attesa di un nodo prima che diventi healthy
var CHECK_TERMINATING_INTERVAL time.Duration = 30 * time.Second     // Ogni quanto effettuare il controllo sulle istanze in terminazione
var ACTIVITY_CACHE_FLUSH_INTERVAL time.Duration = 600 * time.Second // Ogni quanto flushare la cache sulle istanze in terminazione

// Port Settings
var HEARTBEAT_PORT string = ":8888" // Porta su cui il nodo ascolta i segnali da load balancer e registry
var UPDATES_PORT string = ":27001"  // Porta su cui il nodo ascolta l'update mongo da altri nodi
var RPC_PORT string = ":80"         // Porta su cui il nodo ascolta le chiamate RPC
