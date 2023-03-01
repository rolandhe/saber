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
	log.Printf("[Info] "+format, v...)
}

func (logger *DefaultLogger) InfoLn(v ...any) {
	merge := make([]any, 0, len(v)+1)
	merge = append(merge, "[Info] ")
	merge = append(merge, v...)
	log.Println(merge...)
}
