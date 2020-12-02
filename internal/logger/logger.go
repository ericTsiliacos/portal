package logger

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var (
	LogInfo  *log.Logger
	LogError *log.Logger
	LogPath  string
	logFile  *os.File
)

func init() {
	usr, _ := user.Current()
	home := usr.HomeDir
	LogPath = filepath.Join(home, "/.portal/Logs")
	err := os.MkdirAll(LogPath, 0770)
	if err != nil {
		panic(err)
	}

	logFilePath := filepath.Join(LogPath, "/info.log")
	logFile, err = os.Create(logFilePath)
	if err != nil {
		panic(err)
	}

	LogInfo = log.New(logFile, "INFO: ", log.LstdFlags|log.Lshortfile)
	LogError = log.New(logFile, "ERROR: ", log.LstdFlags|log.Lshortfile)
}

func CloseLogOutput() {
	logFile.Close()
}
