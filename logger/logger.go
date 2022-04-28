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
	LogLevelErr = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

const LogLevel = LogLevelInfo

func init() {
	out := os.Stdout
	debugLogger = log.New(out, "[D]", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(out, "[I]", log.Ldate|log.Ltime)
	warnLogger = log.New(out, "[W]", log.Ldate|log.Ltime)
	errLogger = log.New(out, "[E]", log.Ldate|log.Ltime)
}

func Init(w io.Writer) {
	infoLogger.SetOutput(w)
	warnLogger.SetOutput(w)
	errLogger.SetOutput(w)
}

func InitDebug(w io.Writer) {
	debugLogger.SetOutput(w)
}

func Debug(msg string) {
	if LogLevel >= LogLevelDebug {
		debugLogger.Print(msg)
	}
}

func Debugf(format string, v ...interface{}) {
	if LogLevel >= LogLevelDebug {
		debugLogger.Printf(format, v...)
	}
}

func Info(msg string) {
	if LogLevel >= LogLevelInfo {
		infoLogger.Print(msg)
	}
}

func Infof(format string, v ...interface{}) {
	if LogLevel >= LogLevelInfo {
		infoLogger.Printf(format, v...)
	}
}

func Warn(msg string) {
	if LogLevel >= LogLevelWarn {
		warnLogger.Print(msg)
	}
}

func Warnf(format string, v ...interface{}) {
	if LogLevel >= LogLevelWarn {
		warnLogger.Printf(format, v...)
	}
}

func Error(msg string) {
	if LogLevel >= LogLevelErr {
		errLogger.Print(msg)
	}
}

func Errorf(format string, v ...interface{}) {
	if LogLevel >= LogLevelErr {
		errLogger.Printf(format, v...)
	}
}
