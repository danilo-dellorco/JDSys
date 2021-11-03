package utils

import "time"

// MongoDB Settings
var MAX_TIME time.Duration = 10 // 1 Ora
var CSV string = ".csv"
var CLOUD_EXPORT_PATH string = "../localsys/communication/cloud/export/"
var CLOUD_EXPORT_FILE string = CLOUD_EXPORT_PATH + "exported.csv"
var CLOUD_RECEIVE_PATH string = "../localsys/communication/cloud/receive/"
var UPDATES_EXPORT_FILE string = "../localsys/communication/updates/export/exported.csv"
var UPDATES_RECEIVE_FILE string = "../localsys/communication/updates/receive/received.csv"
