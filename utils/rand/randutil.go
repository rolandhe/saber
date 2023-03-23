// Package randutil, rand  tool
//
// Copyright 2023 The saber Authors. All rights reserved.

// Package randutil 使用go linkname技术暴露runtime 私有函数，以提高性能
package randutil

import _ "unsafe"

// FastRandN 0 ~ n 范围的随机数
//
//go:linkname FastRandN runtime.fastrandn
func FastRandN(n uint32) uint32

// FastRand 生成一个64位的随机数
//
//go:linkname FastRand runtime.fastrand64
func FastRand() uint64
