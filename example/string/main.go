package main

import (
	"fmt"
	"github.com/rolandhe/saber/jcomp"
	"github.com/rolandhe/saber/strutils"
)

func main() {
	//javaLength()
	quickString()
}

func javaLength() {
	s := "刘德华 andi lou"
	l, _ := jcomp.JavaStringLen(s)

	fmt.Println(l, len(s), strutils.GetRuneLenOfString(s))
}

func quickString() {
	s := "刘德华 andi lou"
	b := strutils.DetachBytesString(s)
	s1 := strutils.AttachBytesString(b)
	fmt.Println(s1)
}
