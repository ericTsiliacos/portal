package logger

import (
	"fmt"
	"log"
	"os"
	"os/user"
)

var (
	Log *log.Logger
)

func init() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	err := os.MkdirAll(dir+"/Library/Logs/Portal", 0770)
	if err != nil {
		panic(err)
	}

	var logPath = dir + "/Library/Logs/Portal/info.log"
	file, err := os.Create(logPath)
	if err != nil {
		panic(err)
	}

	Log = log.New(file, "INFO: ", log.LstdFlags|log.Lshortfile)
	fmt.Println("LogFile : " + logPath)
}
