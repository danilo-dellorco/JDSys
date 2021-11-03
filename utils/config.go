package utils

import "time"

// MongoDB Settings
var MAX_TIME time.Duration = 10 // 1 Ora
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
var AWS_CRED_PATH string = "/home/danilo/.aws/credentials"
