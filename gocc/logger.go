package gocc

import "log"

var CcLogger = &defaultLogger{}

type Logger interface {
	Info(format string, v ...any)
	InfoLn(v ...any)
}

type defaultLogger struct {
}

func (logger *defaultLogger) Info(format string, v ...any) {
	log.Printf(format, v...)
}

func (logger *defaultLogger) InfoLn(v ...any) {
	log.Println(v...)
}
