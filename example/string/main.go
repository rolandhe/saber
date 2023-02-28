package main

import (
	"fmt"
	"github.com/rolandhe/saber/jcomp"
	"github.com/rolandhe/saber/utils/strutils"
)

func main() {
	//javaLength()
	getChan()
	//quickString()
}

func getChan() {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	close(ch)
	v, c := <-ch
	fmt.Println(v, c)
	v, c = <-ch
	fmt.Println(v, c)
	v, c = <-ch
	fmt.Println(v, c)
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
