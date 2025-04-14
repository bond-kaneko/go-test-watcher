package main

import (
	"testing"
)

func TestReverse(t *testing.T) {
	util := NewStringUtil()
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"world", "dlrow"},
		{"", ""},
		{"a", "a"},
		{"日本語", "語本日"},
	}

	for _, test := range tests {
		result := util.Reverse(test.input)
		if result != test.expected {
			t.Errorf("Reverse(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestCountChars(t *testing.T) {
	util := NewStringUtil()
	tests := []struct {
		input    string
		expected int
	}{
		{"hello", 5},
		{"", 0},
		{"日本語", 3},
	}

	for _, test := range tests {
		result := util.CountChars(test.input)
		if result != test.expected {
			t.Errorf("CountChars(%q) = %d, want %d", test.input, result, test.expected)
		}
	}
}

func TestToUppercase(t *testing.T) {
	util := NewStringUtil()
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "HELLO"},
		{"Hello", "HELLO"},
		{"HELLO", "HELLO"},
		{"hello123", "HELLO123"},
	}

	for _, test := range tests {
		result := util.ToUppercase(test.input)
		if result != test.expected {
			t.Errorf("ToUppercase(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestToLowercase(t *testing.T) {
	util := NewStringUtil()
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello", "hello"},
		{"hello", "hello"},
		{"HELLO123", "hello123"},
	}

	for _, test := range tests {
		result := util.ToLowercase(test.input)
		if result != test.expected {
			t.Errorf("ToLowercase(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestContains(t *testing.T) {
	util := NewStringUtil()
	tests := []struct {
		str      string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", " ", true},
		{"hello world", "xyz", false},
		{"", "", true},
	}

	for _, test := range tests {
		result := util.Contains(test.str, test.substr)
		if result != test.expected {
			t.Errorf("Contains(%q, %q) = %v, want %v", test.str, test.substr, result, test.expected)
		}
	}
}
