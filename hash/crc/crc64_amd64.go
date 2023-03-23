// Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.

// Package crc 通过汇编调用intel _mm_crc32_u64指令来加速crc的计算，
// 要求机器必须支持SSE42指令
package crc

import (
	"runtime"
	"unsafe"
)

const (
	cpuidSSE42 = 1 << 20
)

var withSSE42 bool

//go:noescape
//go:nosplit
func _crc32u64(a, b uint64, result unsafe.Pointer)

// cpuid is implemented in cpu_x86.s.
//
//go:noescape
//go:nosplit
func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)

func init() {
	withSSE42 = _withSSE42()
}

func _withSSE42() bool {
	if runtime.GOARCH != "amd64" {
		return false
	}
	_, _, ecx1, _ := cpuid(1, 0)
	return ecx1&cpuidSSE42 != 0
}

// Crc32u64  通过 _mm_crc32_u64 指令完成crc算法
func Crc32u64(a, b uint64) uint64 {
	var sum uint64
	_crc32u64(a, b, unsafe.Pointer(&sum))
	return sum
}

// WithSSE42 判断当前系统是否支持 SSE42 指令
func WithSSE42() bool {
	return withSSE42
}
