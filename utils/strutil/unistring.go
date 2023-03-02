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

// GetRuneLenOfString 返回一个字符串的unicode字符个数
func GetRuneLenOfString(s string) int {
	return utf8.RuneCountInString(s)
}

// DetachBytesString 直接获取string底层的utf8字节数组并转换成只读的[]byte,
// 注意, 返回的[]byte只能被读取,不能被修改
// 使用[]byte(s)会造成底层Slice复制字符串的内容,造成内存的分配,效率不好,在只读场景中可以使用本函数来提高效率
func DetachBytesString(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sl := &reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}

	return *(*[]byte)(unsafe.Pointer(sl))
}

// AttachBytesString 生成直接挂接[]byte的字符串,与DetachBytesString类似,他会提高效率,
// 使用string([]byte)底层会分配内存,造成效率不高,使用本方法会避免分配内存,
// 在打印输出json输出的内容时会非常有效
func AttachBytesString(b []byte) string {
	sl := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := &reflect.StringHeader{Data: sl.Data, Len: sl.Len}

	return *(*string)(unsafe.Pointer(sh))
}
