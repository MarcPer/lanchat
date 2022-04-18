package logger

import (
	"io"
	"log"
	"os"
)

var (
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errLogger   *log.Logger
)

const (
	logLevelErr = iota
	logLevelWarn
	logLevelInfo
	logLevelDebug
)

const LogLevel = logLevelInfo

func init() {
	out := os.Stdout
	debugLogger = log.New(out, "[D]", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(out, "[I]", log.Ldate|log.Ltime)
	warnLogger = log.New(out, "[W]", log.Ldate|log.Ltime)
	errLogger = log.New(out, "[E]", log.Ldate|log.Ltime)
}

func Init(w io.Writer) {
	debugLogger = log.New(w, "[D]", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(w, "[I]", log.Ldate|log.Ltime)
	warnLogger = log.New(w, "[W]", log.Ldate|log.Ltime)
	errLogger = log.New(w, "[E]", log.Ldate|log.Ltime)
}

func Debug(msg string) {
	if LogLevel >= logLevelDebug {
		debugLogger.Print(msg)
	}
}

func Debugf(format string, v ...interface{}) {
	if LogLevel >= logLevelDebug {
		debugLogger.Printf(format, v...)
	}
}

func Info(msg string) {
	if LogLevel >= logLevelInfo {
		infoLogger.Print(msg)
	}
}

func Infof(format string, v ...interface{}) {
	if LogLevel >= logLevelInfo {
		infoLogger.Printf(format, v...)
	}
}

func Warn(msg string) {
	if LogLevel >= logLevelWarn {
		warnLogger.Print(msg)
	}
}

func Warnf(format string, v ...interface{}) {
	if LogLevel >= logLevelWarn {
		warnLogger.Printf(format, v...)
	}
}

func Error(msg string) {
	if LogLevel >= logLevelErr {
		errLogger.Print(msg)
	}
}

func Errorf(format string, v ...interface{}) {
	if LogLevel >= logLevelErr {
		errLogger.Printf(format, v...)
	}
}
