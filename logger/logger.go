// Package logger
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package logger

import "log"

type Logger interface {
	Info(format string, v ...any)
	InfoLn(v ...any)
}

type DefaultLogger struct {
}

func (logger *DefaultLogger) Info(format string, v ...any) {
	log.Printf(format, v...)
}

func (logger *DefaultLogger) InfoLn(v ...any) {
	log.Println(v...)
}
