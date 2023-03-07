// Package hash, Golang concurrent tools like java juc.
//
// Copyright 2023 The saber Authors. All rights reserved.

package hash

import (
	"fmt"
	"testing"
)

func TestCityHash32Le4(t *testing.T) {
	s := "中"

	hash := CityHash32String(s)
	if hash != 0x94f00d7d {
		t.Error("error")
	}
}

func TestCityHash32From5To12(t *testing.T) {
	s := "中国"

	hash := CityHash32String(s)
	if hash != 0xb342166b {
		t.Error("error")
	}
}

func TestCityHash32From12To24(t *testing.T) {
	s := "中国人民了"

	hash := CityHash32String(s)
	if hash != 0xbc72f21a {
		t.Error("error")
	}
}

func TestCityHash32Ge24(t *testing.T) {
	s := "我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件"

	hash := CityHash32String(s)
	if hash != 0x51020cae {
		t.Error("error")
	}
}

func TestCityHash64Le4(t *testing.T) {
	s := "中"

	hash := CityHash64String(s)
	if hash != 0xac065168c34af7fe {
		t.Error("error")
	}
}

func TestCityHash64From4To8(t *testing.T) {
	s := "中国"

	hash := CityHash64String(s)
	if hash != 0xe4499ba8daad0087 {
		t.Error("error")
	}
}

func TestCityHash64From8To16(t *testing.T) {
	s := "中国人民"

	hash := CityHash64String(s)
	if hash != 0x155f37b6451fcd43 {
		t.Error("error")
	}
}

func TestCityHash64From17To32(t *testing.T) {
	s := "中国人民共和国"

	hash := CityHash64String(s)
	if hash != 0xe0665a552991a515 {
		t.Error("error")
	}
}

func TestCityHash64From33To64(t *testing.T) {
	s := "中国人民共和国站起来了鼓掌"

	hash := CityHash64String(s)
	if hash != 0xaf216c93667b7d75 {
		t.Error("error")
	}
}

func TestCityHash64Ge64(t *testing.T) {
	s := "我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件"

	hash := CityHash64String(s)
	if hash != 0x8db857dd3d4b22e1 {
		t.Error("error")
	}
}

func TestCityHash128Less16(t *testing.T) {
	s := "中国人民"

	hash := CityHash128String(s)
	if hash.low != 0xdb49338f94e45026 || hash.high != 0xb57c012f2ec4ee25 {
		t.Error("error")
	}
}

func TestCityHash128Ge16(t *testing.T) {
	s := "中国人民站起来了胡扯"

	hash := CityHash128String(s)
	if hash.low != 0xd44fe88bbed69110 || hash.high != 0x731d6d9846735ecd {
		t.Error("error")
	}
}

func TestCityHash128Ge16Big(t *testing.T) {
	s := "我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件"

	hash := CityHash128String(s)
	if hash.low != 0x73b63e8cd44766c5 || hash.high != 0x9ed8d2c68d45b293 {
		t.Error("error")
	}
}

func TestCityHashCrc256String(t *testing.T) {
	s := "我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件"

	hash,err := CityHashCrc256String(s)
	fmt.Println(hash,err)
	//if hash.low != 0x73b63e8cd44766c5 || hash.high != 0x9ed8d2c68d45b293 {
	//	t.Error("error")
	//}
}
