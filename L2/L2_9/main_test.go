package main

import "testing"

func TestMultString(t *testing.T) {
	tests := []struct {
		name     string
		inp      string
		expected string
	}{
		{"extra-task", "abc\\4\\5", "abc45"},
		{"empty line", "", ""},
		{"base case", "abc3", "abccc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := multString(tt.inp)
			if err != nil {
				t.Errorf("multString failed with %v", err)
			}
			if result != tt.expected {
				t.Errorf("multString failed with %v", tt.expected)
			}
		})
	}

}
