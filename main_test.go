package main

import (
	"fmt"
	"testing"
)

func TestBasic(t *testing.T) {
	fmt.Println(Get(Req{"ich", "de"}))
}
