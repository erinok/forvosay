package main

import (
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	fmt.Println(GetPronunciations(Req{"ich", "de"}))
}

func ExampleHexsha1() {
	fmt.Println(hexsha1("hello"))
	// Output:
	// aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
}
