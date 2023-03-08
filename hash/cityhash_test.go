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

func TestCityHashCrc256StringLess240(t *testing.T) {
	s := "我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件"

	hash, err := CityHashCrc256String(s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(hash, err)
	if hash[0] != 0x3A7D187FB7F66B24 || hash[1] != 0xF1A46DAFEDEA9745 || hash[2] != 0x59C56C88A14437CD || hash[3] != 0xE266B7432F21E8AF {
		t.Error("error")
	}
}

func TestCityHashCrc256StringGreat240(t *testing.T) {
	s := "我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件Width and precision are measured in units of Unicode code points, that is, runes. (This differs from C's printf where the units are always measured in bytes.) Either or both of the flags may be replaced with the character '*', causing their values to be obtained from the next operand (preceding the one to format), which must be of type int."

	hash, err := CityHashCrc256String(s)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(hash, err)
	if hash[0] != 0x5C0042CE0214D738 || hash[1] != 0x73B87E64358B3FC || hash[2] != 0xC9DD169C8EA1C7B7 || hash[3] != 0x794CF48A69E4E2BA {
		t.Error("error")
	}
}

func TestCityHashCrc256String240(t *testing.T) {
	s := "我们将通过生成一个大的文件的方式来检验各种方法的执行效率因为这种方式在结束的时候需要执行文件Width and precision are measured in units of Unicode code points, that is, runes. (This differs from C"

	hash, err := CityHashCrc256String(s)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(hash, err)
	if hash[0] != 0xFDF2DAF446298495 || hash[1] != 0x8EED1340BE726CEA || hash[2] != 0x63702635D7CE52F9 || hash[3] != 0x616FD421E1014D0C {
		t.Error("error")
	}
}
