package main

import (
	"reflect"
	"testing"
)

func TestSplitList(t *testing.T) {
	got := splitList(" https://one.example,https://two.example, ")
	want := []string{"https://one.example", "https://two.example"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitList() = %#v, want %#v", got, want)
	}
}
