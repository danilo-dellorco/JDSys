package utils

import (
	"fmt"
	"time"
)

func GetTimestamp(message string) {
	dt := time.Now()
	fmt.Println(message+":", dt.Format("01-02-2006 15:04:05.000000000"))
}
