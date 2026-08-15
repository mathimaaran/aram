package main

import "testing"

func TestDefaultCOptimization(t *testing.T) {
	if defaultCOpt != "-O2" {
		t.Fatalf("default C optimization = %q, want -O2", defaultCOpt)
	}
}
