package logger

import (
    "log"
    "os"
)

type prefix func(string) string
var appLog = log.New(os.Stderr, "", log.Ldate | log.Ltime | log.Lshortfile)


func GetLogger() *log.Logger {
    return appLog
}
