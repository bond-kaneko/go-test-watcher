package main

// StringUtil provides simple string manipulation utilities
type StringUtil struct{}

// NewStringUtil creates a new string utility
func NewStringUtil() *StringUtil {
	return &StringUtil{}
}

// Reverse returns the reversed version of the input string
func (s *StringUtil) Reverse(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// CountChars returns the number of characters in the string
func (s *StringUtil) CountChars(input string) int {
	return len([]rune(input))
}

// ToUppercase converts the string to uppercase
func (s *StringUtil) ToUppercase(input string) string {
	result := ""
	for _, char := range input {
		if char >= 'a' && char <= 'z' {
			result += string(char - 32)
		} else {
			result += string(char)
		}
	}
	return result
}

// ToLowercase converts the string to lowercase
func (s *StringUtil) ToLowercase(input string) string {
	result := ""
	for _, char := range input {
		if char >= 'A' && char <= 'Z' {
			result += string(char + 32)
		} else {
			result += string(char)
		}
	}
	return result
}

// Contains checks if a string contains a substring
func (s *StringUtil) Contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
