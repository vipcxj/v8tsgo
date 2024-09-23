package test

import "testing"

func AssertEqual[T comparable](t *testing.T, expected T, real T, prefixMsg string) {
	t.Helper()
	if expected != real {
		t.Errorf("%sexpect %v, but got %v", prefixMsg, expected, real)
	}
}

func MustEqual[T comparable](t *testing.T, expected T, real T, prefixMsg string) {
	t.Helper()
	if expected != real {
		t.Fatalf("%sexpect %v, but got %v", prefixMsg, expected, real)
	}
}