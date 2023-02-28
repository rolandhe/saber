// Package strutils, string tool
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package strutil

import (
	"reflect"
	"unicode/utf8"
	"unsafe"
)

func GetRuneLenOfString(s string) int {
	return utf8.RuneCountInString(s)
}

func DetachBytesString(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sl := &reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}

	return *(*[]byte)(unsafe.Pointer(sl))
}

func AttachBytesString(b []byte) string {
	sl := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := &reflect.StringHeader{Data: sl.Data, Len: sl.Len}

	return *(*string)(unsafe.Pointer(sh))
}
