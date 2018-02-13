package main

import (
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	fmt.Println(GetPronunciations(Req{"ich", "de"}))
}
