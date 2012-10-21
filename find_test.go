// Copyright © 2012 Steve McCoy under the MIT license.

package main

import (
	"testing"
)

func TestFindTests(t *testing.T) {
	tests := []struct{
		code string
		expected []string
	}{
		{ "void test_frob", []string{ "test_frob" } },
		{ "void test_frob(xyz)", []string{ "test_frob" } },
		{ `
void test_cat
void test_dog
int test_robot
void test_hamster
1 + 2 = 3
void test_gopher`, []string{ "test_cat", "test_dog", "test_hamster", "test_gopher" } },
	}

	for _, test := range tests {
		found := findTests([]byte(test.code))
		if !slicesEqual(found, test.expected) {
			t.Errorf("Slices for code [%s] differ: Expected %v, got %v", test.code, test.expected, found)
		}
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
