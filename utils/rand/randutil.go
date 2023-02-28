package randutil

import _ "unsafe"

//go:linkname FastRandN runtime.fastrandn
func FastRandN(n uint32) uint32

//go:linkname FastRand runtime.fastrand64
func FastRand() uint64
