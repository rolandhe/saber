// Package logger
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package logger

import (
	"github.com/rolandhe/saber/utils"
	"log"
	"strconv"
)

const (
	DebugLevel = iota
	InfoLevel
	ErrorLevel
)

// Logger 封装log类似与java slf4j的功能
type Logger interface {
	Debug(format string, v ...any)
	DebugLn(v ...any)

	Info(format string, v ...any)
	InfoLn(v ...any)

	Error(format string, v ...any)
	ErrorLn(v ...any)
}

// NewDefaultLogger 构建Info级别的缺省日志输出实例
func NewDefaultLogger() Logger {
	return NewLoggerWithLevel(InfoLevel)
}

// NewLoggerWithLevel  指定日志级别并生成对应的缺省日志输出实例
func NewLoggerWithLevel(logLevel int) Logger {
	return &defaultLogger{
		logLevel,
	}
}

type defaultLogger struct {
	logLevel int
}

func (logger *defaultLogger) Debug(format string, v ...any) {
	if logger.IsEnableDebug() {
		logOutput("[Debug]", format, v...)
	}
}

func (logger *defaultLogger) DebugLn(v ...any) {
	if logger.IsEnableDebug() {
		logOutputLn("[Debug]", v...)
	}
}

func (logger *defaultLogger) Info(format string, v ...any) {
	if logger.IsEnableInfo() {
		logOutput("[Info]", format, v...)
	}
}

func (logger *defaultLogger) InfoLn(v ...any) {
	if logger.IsEnableInfo() {
		logOutputLn("[Info]", v...)
	}
}

func (logger *defaultLogger) Error(format string, v ...any) {
	if logger.IsEnableError() {
		logOutput("[Error]", format, v...)
	}
}

func (logger *defaultLogger) ErrorLn(v ...any) {
	if logger.IsEnableError() {
		logOutputLn("[Error]", v...)
	}
}

func (logger *defaultLogger) IsEnableDebug() bool {
	return logger.logLevel <= DebugLevel
}

func (logger *defaultLogger) IsEnableInfo() bool {
	return logger.logLevel <= InfoLevel
}

func (logger *defaultLogger) IsEnableError() bool {
	return logger.logLevel <= ErrorLevel
}

func logOutput(prefix string, format string, v ...any) {
	gid, _ := utils.GetGoRoutineId()
	gidSec := " gid=" + strconv.FormatUint(gid, 10)
	log.Printf(prefix+gidSec+" "+format, v...)
}

func logOutputLn(prefix string, v ...any) {
	merge := make([]any, 0, len(v)+1)
	gid, _ := utils.GetGoRoutineId()
	gidSec := " gid=" + strconv.FormatUint(gid, 10)
	merge = append(merge, prefix+gidSec+" ")
	merge = append(merge, v...)
	log.Println(merge...)
}
