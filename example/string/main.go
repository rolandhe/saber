package main

import (
	"fmt"
	"github.com/rolandhe/saber/jcomp"
)

func main() {
	javaLength()
}

func javaLength() {
	s := "刘德华 andi lou"
	l, _ := jcomp.JavaStringLen(s)

	fmt.Println(l, len(s))
}
