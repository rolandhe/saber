// Package randutil, rand  tool
//
// Copyright 2023 The saber Authors. All rights reserved.

package randutil

import _ "unsafe"

// 使用go linkname技术暴露runtime 私有函数，以提高性能

//go:linkname FastRandN runtime.fastrandn
func FastRandN(n uint32) uint32

//go:linkname FastRand runtime.fastrand64
func FastRand() uint64
