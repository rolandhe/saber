// net framework basing tcp, tcp is 4th layer of osi net model
//
// Copyright 2023 The saber Authors. All rights reserved.

package nfour

import "github.com/rolandhe/saber/logger"

// NFourLogger nfour日志输出实例，nfour底层调用该实例来输出日志，它实现了logger.Logger接口，默认调用标准库的log包输出，您可以通过实现logger.Logger接口来实现
// 封装自己的日志输出，封装类的实例赋值给CcLogger全局变量即可
var NFourLogger = logger.NewDefaultLogger()
