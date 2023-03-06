package hash

import (
	"github.com/rolandhe/saber/utils/strutil"
	"math/bits"
	"unsafe"
)

var littleEndian bool

const k0 uint64 = 0xc3a5c85c97cb3127
const k1 uint64 = 0xb492b66fbe98f273
const k2 uint64 = 0x9ae16a3b2f90404f
const kMul uint64 = 0x9ddfea08eb382d69

// Magic numbers for 32-bit hashing.  Copied from Murmur3.
const c1 uint32 = 0xcc9e2d51
const c2 uint32 = 0x1b873593

func init() {
	littleEndian = IsLittleEndian()
}

type Uint128 struct {
	low  uint64
	high uint64
}

func MakeUint128(u uint64, v uint64) *Uint128 {
	return &Uint128{
		u, v,
	}
}

func IsLittleEndian() bool {
	n := 0x1234
	f := *((*byte)(unsafe.Pointer(&n)))
	return (f ^ 0x34) == 0
}

func fetch64(data []byte) uint64 {
	v := uint64(data[0])
	v |= uint64(data[1]) << 8
	v |= uint64(data[2]) << 16
	v |= uint64(data[3]) << 24
	v |= uint64(data[4]) << 32
	v |= uint64(data[5]) << 40
	v |= uint64(data[5]) << 40
	v |= uint64(data[6]) << 48
	v |= uint64(data[7]) << 56
	if littleEndian {
		return v
	}

	return bits.Reverse64(v)
}

func fetch32(data []byte) uint32 {
	v := uint32(data[0])
	v |= uint32(data[1]) << 8
	v |= uint32(data[2]) << 16
	v |= uint32(data[3]) << 24

	if littleEndian {
		return v
	}

	return bits.Reverse32(v)
}

func fmix(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

func rotate32(val uint32, shift int) uint32 {
	// Avoid shifting by 32: doing so yields an undefined result.
	if shift == 0 {
		return val
	}
	return (val >> shift) | (val << (32 - shift))
}

func mur(a uint32, h uint32) uint32 {
	// Helper from Murmur3 for combining two 32-bit values.
	a *= c1
	a = rotate32(a, 17)
	a *= c2
	h ^= a
	h = rotate32(h, 19)
	return h*5 + 0xe6546b64
}

func hash32Len13to24(s []byte, len uint) uint32 {
	a := fetch32(s[len>>1-4:])
	b := fetch32(s[4:])
	c := fetch32(s[len-8:])
	d := fetch32(s[len>>1:])
	e := fetch32(s)
	f := fetch32(s[len-4:])
	h := uint32(len)
	return fmix(mur(f, mur(e, mur(d, mur(c, mur(b, mur(a, h)))))))
}

func hash32Len0to4(s []byte, len uint) uint32 {
	b := uint32(0)
	c := uint32(9)
	for i := uint(0); i < len; i++ {
		v := int8(s[i])
		b = b*c1 + uint32(v)
		c ^= b
	}
	return fmix(mur(b, mur(uint32(len), c)))
}

func hash32Len5to12(s []byte, len uint) uint32 {
	a := uint32(len)
	b := a * 5
	c := uint32(9)
	d := b
	a += fetch32(s)
	b += fetch32(s[len-4:])
	pos := (len >> 1) & 4
	c += fetch32(s[pos:])
	return fmix(mur(c, mur(b, mur(a, d))))
}

func CityHash32(str string) uint32 {
	s := []byte(str)
	length := uint(len(str))
	if length <= 24 {
		if length <= 12 {
			if length <= 4 {
				return hash32Len0to4(s, length)
			} else {
				return hash32Len5to12(s, length)
			}
		} else {
			return hash32Len13to24(s, length)
		}
	}

	// length > 24
	h := uint32(length)
	g := c1 * h
	f := g
	a0 := rotate32(fetch32(s[length-4:])*c1, 17) * c2
	a1 := rotate32(fetch32(s[length-8:])*c1, 17) * c2
	a2 := rotate32(fetch32(s[length-16:])*c1, 17) * c2
	a3 := rotate32(fetch32(s[length-12:])*c1, 17) * c2
	a4 := rotate32(fetch32(s[length-20:])*c1, 17) * c2
	h ^= a0
	h = rotate32(h, 19)
	h = h*5 + 0xe6546b64
	h ^= a2
	h = rotate32(h, 19)
	h = h*5 + 0xe6546b64
	g ^= a1
	g = rotate32(g, 19)
	g = g*5 + 0xe6546b64
	g ^= a3
	g = rotate32(g, 19)
	g = g*5 + 0xe6546b64
	f += a4
	f = rotate32(f, 19)
	f = f*5 + 0xe6546b64
	iters := (length - 1) / 20
	for {
		a0 := rotate32(fetch32(s)*c1, 17) * c2
		a1 := fetch32(s[4:])
		a2 := rotate32(fetch32(s[8:])*c1, 17) * c2
		a3 := rotate32(fetch32(s[12:])*c1, 17) * c2
		a4 = fetch32(s[16:])
		h ^= a0
		h = rotate32(h, 18)
		h = h*5 + 0xe6546b64
		f += a1
		f = rotate32(f, 19)
		f = f * c1
		g += a2
		g = rotate32(g, 18)
		g = g*5 + 0xe6546b64
		h ^= a3 + a1
		h = rotate32(h, 19)
		h = h*5 + 0xe6546b64
		g ^= a4
		g = bits.Reverse32(g) * 5
		h += a4 * 5
		h = bits.Reverse32(h)
		f += a0
		f, h, g = g, f, h
		s = s[20:]
		iters--
		if iters == 0 {
			break
		}
	}

	g = rotate32(g, 11) * c1
	g = rotate32(g, 17) * c1
	f = rotate32(f, 11) * c1
	f = rotate32(f, 17) * c1
	h = rotate32(h+g, 19)
	h = h*5 + 0xe6546b64
	h = rotate32(h, 17) * c1
	h = rotate32(h+f, 19)
	h = h*5 + 0xe6546b64
	h = rotate32(h, 17) * c1
	return h
}

func rotate64(val uint64, shift int) uint64 {
	// Avoid shifting by 64: doing so yields an undefined result.
	if shift == 0 {
		return val
	}
	return (val >> shift) | (val << (64 - shift))
}

func shiftMix(val uint64) uint64 {
	return val ^ (val >> 47)
}

func hash128to64(u uint64, v uint64) uint64 {
	// Murmur-inspired hashing.
	a := (u ^ v) * kMul
	a ^= a >> 47
	b := (v ^ a) * kMul
	b ^= b >> 47
	b *= kMul
	return b
}

func hashLen16(u uint64, v uint64) uint64 {
	return hash128to64(u, v)
}

func hashLen16WithMul(u uint64, v uint64, mul uint64) uint64 {
	// Murmur-inspired hashing.
	a := (u ^ v) * mul
	a ^= a >> 47
	b := (v ^ a) * mul
	b ^= b >> 47
	b *= mul
	return b
}

func hashLen0to16(s []byte, len uint) uint64 {
	if len >= 8 {
		mul := k2 + uint64(len)*2
		a := fetch64(s) + k2
		b := fetch64(s[len-8:])
		c := rotate64(b, 37)*mul + a
		d := (rotate64(a, 25) + b) * mul
		return hashLen16WithMul(c, d, mul)
	}
	if len >= 4 {
		mul := k2 + uint64(len)*2
		a := uint64(fetch32(s))
		return hashLen16WithMul(uint64(len)+(a<<3), uint64(fetch32(s[len-4:])), mul)
	}
	if len > 0 {
		a := s[0]
		b := s[len>>1]
		c := s[len-1]
		y := uint32(a) + (uint32(b) << 8)
		z := uint32(len) + (uint32(c) << 2)
		return shiftMix(uint64(y)*k2^uint64(z)*k0) * k2
	}
	return k2
}

// This probably works well for 16-byte strings as well, but it may be over kill
// in that case.
func hashLen17to32(s []byte, len uint) uint64 {
	mul := k2 + uint64(len)*2
	a := fetch64(s) * k1
	b := fetch64(s[8:])
	c := fetch64(s[len-8:]) * mul
	d := fetch64(s[len-16:]) * k2
	return hashLen16WithMul(rotate64(a+b, 43)+rotate64(c, 30)+d,
		a+rotate64(b+k2, 18)+c, mul)
}

// Return a 16-byte hash for 48 bytes.  Quick and dirty.
// Callers do best to use "random-looking" values for a and b.
func weakHashLen32WithSeedsBaseNumber(
	w uint64, x uint64, y uint64, z uint64, a uint64, b uint64) *Uint128 {
	a += w
	b = rotate64(b+a+z, 21)
	c := a
	a += x
	a += y
	b += rotate64(a, 44)
	return MakeUint128(a+z, b+c)
}

// Return a 16-byte hash for s[0] ... s[31], a, and b.  Quick and dirty.
func weakHashLen32WithSeeds(
	s []byte, a uint64, b uint64) *Uint128 {
	return weakHashLen32WithSeedsBaseNumber(fetch64(s),
		fetch64(s[8:]),
		fetch64(s[16:]),
		fetch64(s[24:]),
		a,
		b)
}

func hashLen33to64(s []byte, length uint) uint64 {
	mul := k2 + uint64(length)*2
	a := fetch64(s) * k2
	b := fetch64(s[8:])
	c := fetch64(s[length-24:])
	d := fetch64(s[length-32:])
	e := fetch64(s[16:]) * k2
	f := fetch64(s[:24]) * 9
	g := fetch64(s[length-8:])
	h := fetch64(s[length-16:]) * mul
	u := rotate64(a+g, 43) + (rotate64(b, 30)+c)*9
	v := ((a + g) ^ d) + f + 1
	w := bits.Reverse64((u+v)*mul) + h
	x := rotate64(e+f, 42) + c
	y := (bits.Reverse64((v+w)*mul) + g) * mul
	z := e + f + c
	a = bits.Reverse64((x+z)*mul+y) + b
	b = shiftMix((z+a)*mul+d+h) * mul
	return b + x
}

func CityHash64(str string) uint64 {
	s := []byte(str)
	length := uint(len(str))
	if length <= 32 {
		if length <= 16 {
			return hashLen0to16(s, length)
		} else {
			return hashLen17to32(s, length)
		}
	} else if length <= 64 {
		return hashLen33to64(s, length)
	}

	// For strings over 64 bytes we hash the end first, and then as we
	// loop we keep 56 bytes of state: v, w, x, y, and z.
	x := fetch64(s[length-40:])
	y := fetch64(s[length-16:]) + fetch64(s[length-56:])
	z := hashLen16(fetch64(s[length-48:])+uint64(length), fetch64(s[length-24:]))
	v := weakHashLen32WithSeeds(s[length-64:], uint64(length), z)
	w := weakHashLen32WithSeeds(s[length-32:], y+k1, x)
	x = x*k1 + fetch64(s)

	// Decrease length to the nearest multiple of 64, and operate on 64-byte chunks.
	slen := int(length)
	slen = (slen - 1) & ^63
	for {
		x = rotate64(x+y+v.low+fetch64(s[8:]), 37) * k1
		y = rotate64(y+v.high+fetch64(s[48:]), 42) * k1
		x ^= w.high
		y += v.low + fetch64(s[40:])
		z = rotate64(z+w.low, 33) * k1
		v = weakHashLen32WithSeeds(s, v.high*k1, x+w.low)
		w = weakHashLen32WithSeeds(s[32:], z+w.high, y+fetch64(s[16:]))
		z, x = x, z
		s = s[64:]
		slen -= 64
		if slen == 0 {
			break
		}
	}

	return hashLen16(hashLen16(v.low, w.low)+shiftMix(y)*k1+z,
		hashLen16(v.high, w.high)+x)
}

func CityHash64WithSeed(str string, seed uint64) uint64 {
	return cityHash64WithTwoSeeds(str, k2, seed)
}

func cityHash64WithTwoSeeds(str string, seed0 uint64, seed1 uint64) uint64 {
	return hashLen16(CityHash64(str)-seed0, seed1)
}

// cityMurmur  A subroutine for CityHash128().  Returns a decent 128-bit hash for strings
// of any length representable in signed long.  Based on City and Murmur.
func cityMurmur(s []byte, len uint, seed *Uint128) *Uint128 {
	a := seed.low
	b := seed.high
	c := uint64(0)
	d := uint64(0)
	if len <= 16 {
		a = shiftMix(a*k1) * k1
		c = b*k1 + hashLen0to16(s, len)
		cv := c
		if len >= 8 {
			cv = fetch64(s)
		}
		d = shiftMix(a + cv)
	} else {
		c = hashLen16(fetch64(s[len-8:])+k1, a)
		d = hashLen16(b+uint64(len), c+fetch64(s[len-16:]))
		a += d
		// len > 16 here, so do...while is safe
		for {
			a ^= shiftMix(fetch64(s)*k1) * k1
			a *= k1
			b ^= a
			c ^= shiftMix(fetch64(s[8:])*k1) * k1
			c *= k1
			d ^= c
			s = s[16:]
			len -= 16
			if len <= 16 {
				break
			}
		}
	}
	a = hashLen16(a, c)
	b = hashLen16(d, b)
	return MakeUint128(a^b, hashLen16(b, a))
}

func cityHash128WithSeedCore(s []byte, length uint, seed *Uint128) *Uint128 {
	if length < 128 {
		return cityMurmur(s, length, seed)
	}

	// We expect length >= 128 to be the common case.  Keep 56 bytes of state:
	// v, w, x, y, and z.
	var v Uint128
	var w Uint128
	x := seed.low
	y := seed.high
	z := uint64(length) * k1
	v.low = rotate64(y^k1, 49)*k1 + fetch64(s)
	v.high = rotate64(v.low, 42)*k1 + fetch64(s[8:])
	w.low = rotate64(y+z, 35)*k1 + x
	w.high = rotate64(x+fetch64(s[88:]), 53) * k1

	// This is the same inner loop as CityHash64(), manually unrolled.
	for {
		x = rotate64(x+y+v.low+fetch64(s[8:]), 37) * k1
		y = rotate64(y+v.high+fetch64(s[48:]), 42) * k1
		x ^= w.high
		y += v.low + fetch64(s[40:])
		z = rotate64(z+w.low, 33) * k1
		v = *weakHashLen32WithSeeds(s, v.high*k1, x+w.low)
		w = *weakHashLen32WithSeeds(s[32:], z+w.high, y+fetch64(s[16:]))
		z, x = x, z
		s = s[64:]
		x = rotate64(x+y+v.low+fetch64(s[8:]), 37) * k1
		y = rotate64(y+v.high+fetch64(s[48:]), 42) * k1
		x ^= w.high
		y += v.low + fetch64(s[40:])
		z = rotate64(z+w.low, 33) * k1
		v = *weakHashLen32WithSeeds(s, v.high*k1, x+w.low)
		w = *weakHashLen32WithSeeds(s[32:], z+w.high, y+fetch64(s[16:]))
		z, x = x, z
		s = s[64:]
		length -= 128
		if length < 128 {
			break
		}
	}
	x += rotate64(v.low+z, 49) * k0
	y = y*k0 + rotate64(w.high, 37)
	z = z*k0 + rotate64(w.low, 27)
	w.low *= 9
	v.low *= k0
	// If 0 < length < 128, hash up to 4 chunks of 32 bytes each from the end of s.
	for tailDone := uint(0); tailDone < length; {
		tailDone += 32
		y = rotate64(x+y, 42)*k0 + v.high
		w.low += fetch64(s[length-tailDone+16:])
		x = x*k0 + w.low
		z += w.high + fetch64(s[length-tailDone:])
		w.high += v.low
		v = *weakHashLen32WithSeeds(s[length-tailDone:], v.low+z, v.high)
		v.low *= k0
	}
	// At this point our 56 bytes of state should contain more than
	// enough information for a strong 128-bit hash.  We use two
	// different 56-byte-to-8-byte hashes to get a 16-byte final result.
	x = hashLen16(x, v.low)
	y = hashLen16(y+z, w.low)
	return MakeUint128(hashLen16(x+v.high, w.high)+y,
		hashLen16(x+w.high, y+v.high))
}

func CityHash128(str string) *Uint128 {
	length := uint(len(str))
	if length >= 16 {
		s := strutil.DetachBytesString(str)
		seed := MakeUint128(fetch64(s), fetch64(s[8:])+k0)

		return cityHash128WithSeedCore(s[16:], length-16, seed)
	}
	return CityHash128WithSeed(str, MakeUint128(k0, k1))
}

func CityHash128WithSeed(str string, seed *Uint128) *Uint128 {
	s := strutil.DetachBytesString(str)
	length := uint(len(str))
	return cityHash128WithSeedCore(s, length, seed)
}
