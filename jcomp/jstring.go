// Package jcomp, compatible with java string.
//
// Copyright 2023 The saber Authors. All rights reserved.
//

package jcomp

import (
	"errors"
	"fmt"
)

const (
	MinHighSurrogate          = rune(55296)
	MaxHighSurrogate          = rune(56319)
	MinLowSurrogate           = rune(56320)
	MaxLowSurrogate           = rune(57343)
	MinSupplementaryCodePoint = rune(0x010000)
	MinCodePoint              = rune(0x000000)
	MaxCodePoint              = rune(0x10FFFF)
)

type Char uint16
type CodePoint rune

func JavaStringLen(s string) (int, error) {
	_, l, err := javaStringChars(s, false)
	return l, err
}

func JavaSubStringToEnd(s string, start int) (string, error) {
	return JavaSubString(s, start, -1)
}

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

func JavaToChars(cp rune) ([]Char, error) {
	ret := []Char{0, 0}

	l, err := toJavaChars(cp, ret)
	if err != nil {
		return nil, err
	}

	return ret[:l], nil
}

func JavaCharCount(cp rune) int {
	if cp >= MinSupplementaryCodePoint {
		return 2
	}
	return 1
}

func JavaCodePointAt(a []Char, index int) rune {
	return codePointAtImpl(a, index, len(a))
}

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
