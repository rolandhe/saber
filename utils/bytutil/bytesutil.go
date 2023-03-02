// Package bytutil, byte operate tool
//
// Copyright 2023 The saber Authors. All rights reserved.

package bytutil

import "errors"

const mask = 0xff

// ToInt32 转换4个字节为int32
func ToInt32(buf []byte) (int32, error) {
	if len(buf) < 4 {
		return 0, errors.New("int32 need at least 4 bytes")
	}
	ret := int32(buf[0])

	ret |= int32(buf[1]) << 8
	ret |= int32(buf[2]) << 16
	ret |= int32(buf[3]) << 24
	return ret, nil
}

func ToUint32(buf []byte) (uint32, error) {
	ret, err := ToInt32(buf)
	return uint32(ret), err
}

func ToInt64(buf []byte) (int64, error) {
	if len(buf) < 8 {
		return 0, errors.New("int32 need at least 8 bytes")
	}
	ret := int64(buf[0])

	ret |= int64(buf[1]) << 8
	ret |= int64(buf[2]) << 16
	ret |= int64(buf[3]) << 24
	ret |= int64(buf[4]) << 32
	ret |= int64(buf[5]) << 40
	ret |= int64(buf[6]) << 48
	ret |= int64(buf[7]) << 56
	return ret, nil
}

// ToUint64 转换8个字节为unit64
func ToUint64(buf []byte) (uint64, error) {
	ret, err := ToInt64(buf)
	return uint64(ret), err
}

func Int32ToBytes(v int32) []byte {
	return Uint32ToBytes(uint32(v))
}

func Uint32ToBytes(uv uint32) []byte {
	buf := make([]byte, 4, 4)
	buf[0] = byte(uv & mask)
	buf[1] = byte(uv >> 8 & mask)
	buf[2] = byte(uv >> 16 & mask)
	buf[3] = byte(uv >> 24 & mask)
	return buf
}

func Int64ToBytes(v int64) []byte {
	return Uint64ToBytes(uint64(v))
}

func Uint64ToBytes(uv uint64) []byte {
	buf := make([]byte, 8)

	buf[0] = byte(uv & mask)
	buf[1] = byte(uv >> 8 & mask)
	buf[2] = byte(uv >> 16 & mask)
	buf[3] = byte(uv >> 24 & mask)
	buf[4] = byte(uv >> 32 & mask)
	buf[5] = byte(uv >> 40 & mask)
	buf[6] = byte(uv >> 48 & mask)
	buf[7] = byte(uv >> 56 & mask)

	return buf
}
