package main

import "testing"

func TestBasic(t *testing.T) {
	// This test always passes - just verifies the package compiles
	if 1 != 1 {
		t.Error("Math is broken")
	}
}
