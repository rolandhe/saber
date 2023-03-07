package crc

import (
	"unsafe"
)

const (
	cpuid_SSE42 = 1 << 20
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
	_, _, ecx1, _ := cpuid(1, 0)
	return ecx1&cpuid_SSE42 != 0
}

func Crc32u64(a, b uint64) uint64 {
	var sum uint64
	_crc32u64(a, b, unsafe.Pointer(&sum))
	return sum
}

func WithSSE42() bool {
	return withSSE42
}
