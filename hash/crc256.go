// Package hash, Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.

package hash

import (
	"errors"
	"github.com/rolandhe/saber/hash/crc"
	"github.com/rolandhe/saber/utils/strutil"
)

var notSupportError = errors.New("not support sse42")

type param256 struct {
	a uint64
	b uint64
	c uint64
	d uint64
	e uint64
	f uint64
	g uint64
	h uint64
	x uint64
	y uint64
	z uint64
}

func cityHashCrc256Long(s []byte, len uint, seed uint32) ([]uint64, error) {
	if !crc.WithSSE42() {
		return nil, notSupportError
	}
	result := make([]uint64, 4)
	var param param256
	param.a = fetch64(s[56:]) + k0
	param.b = fetch64(s[96:]) + k0
	result[0] = hashLen16(param.b, uint64(len))
	param.c = result[0]
	result[1] = fetch64(s[120:])*k0 + uint64(len)
	param.d = result[1]
	param.e = fetch64(s[184:]) + uint64(seed)
	param.f = 0
	param.g = 0
	param.h = param.c + param.d
	param.x = uint64(seed)
	param.y = 0
	param.z = 0

	iters := len / 240
	len -= iters * 240

	for {
		chunk(0, &param, s)
		permute3(&param.a, &param.h, &param.c)
		s = s[40:]

		chunk(33, &param, s)
		permute3(&param.a, &param.h, &param.f)
		s = s[40:]

		chunk(0, &param, s)
		permute3(&param.b, &param.h, &param.f)
		s = s[40:]

		chunk(42, &param, s)
		permute3(&param.b, &param.h, &param.d)
		s = s[40:]

		chunk(0, &param, s)
		permute3(&param.b, &param.h, &param.e)
		s = s[40:]

		chunk(33, &param, s)
		permute3(&param.a, &param.h, &param.e)
		s = s[40:]

		iters--
		if iters <= 0 {
			break
		}
	}

	for len >= 40 {
		chunk(29, &param, s)
		s = s[40:]

		param.e ^= rotate64(param.a, 20)
		param.h += rotate64(param.b, 30)
		param.g ^= rotate64(param.c, 40)
		param.f += rotate64(param.d, 34)
		permute3(&param.c, &param.h, &param.g)
		len -= 40
	}
	if len > 0 {
		s = s[len-40:]
		chunk(33, &param, s)
		s = s[40:]
		param.e ^= rotate64(param.a, 43)
		param.h += rotate64(param.b, 42)
		param.g ^= rotate64(param.c, 41)
		param.f += rotate64(param.d, 40)
	}
	result[0] ^= param.h
	result[1] ^= param.g

	param.g += param.h
	param.a = hashLen16(param.a, param.g+param.z)

	param.x += param.y << 32
	param.b += param.x
	param.c = hashLen16(param.c, param.z) + param.h
	param.d = hashLen16(param.d, param.e+result[0])
	param.g += param.e
	param.h += hashLen16(param.x, param.f)
	param.e = hashLen16(param.a, param.d) + param.g
	param.z = hashLen16(param.b, param.c) + param.a
	param.y = hashLen16(param.g, param.h) + param.c
	result[0] = param.e + param.z + param.y + param.x
	param.a = shiftMix((param.a+param.y)*k0)*k0 + param.b
	result[1] += param.a + result[0]
	param.a = shiftMix(param.a*k0)*k0 + param.c
	result[2] = param.a + result[1]
	param.a = shiftMix((param.a+param.e)*k0) * k0
	result[3] = param.a + result[2]
	return result, nil
}

// CityHashCrc256String 计算指定字符串的 256 位hash
// 256位hash使用 4个 int64 返回
func CityHashCrc256String(str string) ([]uint64, error) {
	s := strutil.DetachBytesString(str)
	length := uint(len(str))
	return CityHashCrc256(s, length)
}

// CityHashCrc256 计算指定二进制数组的 256 位hash
// 256位hash使用 4个 int64 返回
func CityHashCrc256(s []byte, len uint) ([]uint64, error) {
	if len >= 240 {
		return cityHashCrc256Long(s, len, 0)
	} else {
		return cityHashCrc256Short(s, len)
	}
}

func CityHashCrc128WithSeedString(str string, seed *Uint128) (*Uint128, error) {
	s := strutil.DetachBytesString(str)
	length := uint(len(str))
	return CityHashCrc128WithSeed(s, length, seed)
}

func CityHashCrc128WithSeed(s []byte, length uint, seed *Uint128) (*Uint128, error) {
	if length <= 900 {
		return CityHash128WithSeed(s, length, seed), nil
	} else {
		result, err := CityHashCrc256(s, length)
		if err != nil {
			return nil, err
		}
		u := seed.high + result[0]
		v := seed.low + result[1]
		return MakeUint128(hashLen16(u, v+result[2]), hashLen16(rotate64(v, 32), u*k0+result[3])), nil
	}
}

func CityHashCrc128String(str string) (*Uint128, error) {
	s := strutil.DetachBytesString(str)
	length := uint(len(str))
	return CityHashCrc128(s, length)
}

func CityHashCrc128(s []byte, length uint) (*Uint128, error) {
	if length <= 900 {
		return CityHash128(s, length), nil
	} else {
		result, err := CityHashCrc256(s, length)
		if err != nil {
			return nil, err
		}
		return MakeUint128(result[2], result[3]), nil
	}
}

func chunk(r int, param *param256, s []byte) {
	param.x, param.z, param.y = param.y, param.x, param.z
	param.b += fetch64(s)
	param.c += fetch64(s[8:])
	param.d += fetch64(s[16:])
	param.e += fetch64(s[24:])
	param.f += fetch64(s[32:])
	param.a += param.b
	param.h += param.f
	param.b += param.c
	param.f += param.d
	param.g += param.e
	param.e += param.z
	param.g += param.x
	param.z = crc.Crc32u64(param.z, param.b+param.g)
	param.y = crc.Crc32u64(param.y, param.e+param.h)
	param.x = crc.Crc32u64(param.x, param.f+param.a)
	param.e = rotate64(param.e, r)
	param.c += param.e
}

func permute3(a *uint64, b *uint64, c *uint64) {
	*a, *b, *c = *c, *a, *b
}

func cityHashCrc256Short(s []byte, len uint) ([]uint64, error) {
	data := make([]byte, 240)
	copy(data, s)
	return cityHashCrc256Long(data, 240, ^uint32(len))
}
