package logger

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var (
	LogInfo     *log.Logger
	LogError    *log.Logger
	LogFilePath string
	logFile     *os.File
)

func init() {
	usr, _ := user.Current()
	home := usr.HomeDir
	logPath := filepath.Join(home, "/.portal/Logs")
	err := os.MkdirAll(logPath, 0770)
	if err != nil {
		panic(err)
	}

	LogFilePath = filepath.Join(logPath, "/info.log")
	logFile, err = os.Create(LogFilePath)
	if err != nil {
		panic(err)
	}

	LogInfo = log.New(logFile, "INFO: ", log.LstdFlags|log.Lshortfile)
	LogError = log.New(logFile, "ERROR: ", log.LstdFlags|log.Lshortfile)
}

func CloseLogOutput() {
	logFile.Close()
}
