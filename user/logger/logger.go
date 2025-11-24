package logger

import (
	"log"
	"os"
	"io"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func init() {
	// Файл логов
	file, err := os.OpenFile("mytorrent.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("cannot open logfile: ", err)
	}

	// Логи INFO — в файл и stdout
	infoLogger = log.New(
		io.MultiWriter(file),
		"[INFO] ",
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile,
	)

	// Логи ERROR — в файл и stdout
	errorLogger = log.New(
		io.MultiWriter(file),
		"[ERROR] ",
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile,
	)
}

// ---- Публичные функции ----

func Info(msg string) {
	infoLogger.Println(msg)
}

func Infof(format string, args ...any) {
	infoLogger.Printf(format, args...)
}

func Error(msg string) {
	errorLogger.Println(msg)
}

func Errorf(format string, args ...any) {
	errorLogger.Printf(format, args...)
}
