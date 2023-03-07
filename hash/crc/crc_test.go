package crc

import (
	"fmt"
	"testing"
)

func TestCrc64(t *testing.T) {
	a := uint64(10002)
	b := uint64(20002)
	c := Crc32u64(a, b)

	fmt.Println(c)
}
