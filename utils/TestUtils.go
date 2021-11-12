package utils

import (
	"time"
)

func GetTimestamp() time.Time {
	return time.Now()
}

func FormatTime(t time.Time) string {
	return t.Format("01-02-2006 15:04:05.000000000")
}
