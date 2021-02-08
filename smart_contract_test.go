package main

import (
	"fmt"
	"testing"
)

type item struct {
	Value string `json:"value"`
	Id    int    `json:"id"`
}

func TestSmartContract_ReadUserProfile(t *testing.T) {
	a := make(map[string]bool)
	fmt.Println(a["a"])
}
