// Package jcomp, compatible with java string.
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package jcomp

import (
	"errors"
	"fmt"
)

// 兼容java String属性的工具,包括:
// 1. java String的Length
// 2. java String substring方法的功能
// 3. 转换成兼容java char的数组
// 4. java  Character类的功能

const (
	MinHighSurrogate          = rune(55296)
	MaxHighSurrogate          = rune(56319)
	MinLowSurrogate           = rune(56320)
	MaxLowSurrogate           = rune(57343)
	MinSupplementaryCodePoint = rune(0x010000)
	MinCodePoint              = rune(0x000000)
	MaxCodePoint              = rune(0x10FFFF)
)

// Char 对标java char
type Char uint16

// CodePoint unicode字符集码位
type CodePoint rune

func JavaStringLen(s string) (int, error) {
	_, l, err := javaStringChars(s, false)
	return l, err
}

func JavaSubStringToEnd(s string, start int) (string, error) {
	return JavaSubString(s, start, -1)
}

// JavaSubString 生成子串,[start,end)
func JavaSubString(s string, start int, end int) (string, error) {
	content, l, err := javaStringChars(s, true)
	if err != nil {
		return "", err
	}
	if start < 0 && end > l {
		return "", errors.New("exceed strutils range")
	}
	if end == -1 {
		end = l
	}

	if start == end {
		return "", nil
	}
	if isLowSurrogate(content[start]) {
		return "", errors.New("start pos is invalid character")
	}

	if isHighSurrogate(content[end-1]) {
		return "", errors.New("end pos is invalid character")
	}

	var retRunes []rune

	for i := start; i < end; i++ {
		rv := codePointAtImpl(content, i, end-i)

		retRunes = append(retRunes, rv)
		if JavaCharCount(rv) == 2 {
			i++
		}
	}
	return string(retRunes), nil
}

// JavaToChars 转换一个rune(即unicode codepoint)为2个Char(即uint)
// 兼容java Character.toChars(int codePoint)
func JavaToChars(cp rune) ([]Char, error) {
	ret := []Char{0, 0}

	l, err := toJavaChars(cp, ret)
	if err != nil {
		return nil, err
	}

	return ret[:l], nil
}

// JavaCharCount 计算一个rune需要几个Char组成
// 兼容java Character.charCount(int codePoint)
func JavaCharCount(cp rune) int {
	if cp >= MinSupplementaryCodePoint {
		return 2
	}
	return 1
}

// JavaCodePointAt 转换Char数组中指定位置的字符所对标的codepoint
// 兼容java Character.codePointAt
func JavaCodePointAt(a []Char, index int) rune {
	return codePointAtImpl(a, index, len(a))
}

// JavaCodePoint 转换Char数组中首位置的字符所对标的codepoint
// 兼容java Character.codePoint
func JavaCodePoint(a []Char) rune {
	return JavaCodePointAt(a, 0)
}

func JavaToCodePoint(high Char, low Char) rune {
	return rune(high)<<10 + rune(low) + MinSupplementaryCodePoint - MinHighSurrogate<<10 - MinLowSurrogate
}

func isBmpCodePoint(cp rune) bool {
	return cp>>16 == 0
}

func isValidCodePoint(cp rune) bool {
	plane := cp >> 16
	return plane < ((MaxCodePoint + 1) >> 16)
}

func lowSurrogate(cp rune) Char {
	return Char(cp&0x3ff + MinLowSurrogate)
}
func highSurrogate(cp rune) Char {
	return Char(cp>>10 + MinHighSurrogate - MinSupplementaryCodePoint>>10)
}

func toSurrogates(cp rune, dst []Char, index int) {
	// We write elements "backwards" to guarantee all-or-nothing
	dst[index+1] = lowSurrogate(cp)
	dst[index] = highSurrogate(cp)
}

func isHighSurrogate(ch Char) bool {
	cv := rune(ch)
	return cv >= MinHighSurrogate && cv < (MaxHighSurrogate+1)
}

func isLowSurrogate(ch Char) bool {
	cv := rune(ch)
	return cv >= MinLowSurrogate && cv < (MaxLowSurrogate+1)
}

func codePointAtImpl(a []Char, index int, limit int) rune {
	c1 := a[index]
	if isHighSurrogate(c1) && index+1 < limit {
		index++
		c2 := a[index]
		if isLowSurrogate(c2) {
			return JavaToCodePoint(c1, c2)
		}
	}
	return rune(c1)
}

func toJavaChars(cp rune, buf []Char) (int, error) {
	if isBmpCodePoint(cp) {
		buf[0] = Char(cp)
		return 1, nil
	} else if isValidCodePoint(cp) {
		toSurrogates(cp, buf, 0)
		return 2, nil
	} else {
		return 0, errors.New(fmt.Sprintf("Not a valid Unicode code point: 0x%X", cp))
	}
}

func javaStringChars(s string, withContent bool) ([]Char, int, error) {
	var ret []Char
	runes := []rune(s)
	l := 0
	buf := []Char{0, 0}
	for _, r := range runes {
		size, err := toJavaChars(r, buf)
		if err != nil {
			return nil, 0, err
		}
		l += size
		if !withContent {
			continue
		}
		if size == 1 {
			ret = append(ret, buf[0])
		} else {
			ret = append(ret, buf...)
		}
	}
	return ret, l, nil
}
