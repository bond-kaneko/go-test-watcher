package main

import "testing"

func TestCalculator(t *testing.T) {
	// Create a new calculator
	calc := NewCalculator()

	// Test initial memory value
	if mem := calc.Memory(); mem != 0 {
		t.Errorf("Initial memory value should be 0, got %f", mem)
	}

	// Test addition
	if result := calc.Add(5); result != 5 {
		t.Errorf("Add(5) should return 5, got %f", result)
	}

	// Test subtraction
	if result := calc.Subtract(2); result != 3 {
		t.Errorf("Subtract(2) should return 3, got %f", result)
	}

	// Test multiplication
	if result := calc.Multiply(3); result != 9 {
		t.Errorf("Multiply(3) should return 9, got %f", result)
	}

	// Test division
	if result := calc.Divide(3); result != 3 {
		t.Errorf("Divide(3) should return 3, got %f", result)
	}

	// Test division by zero
	if result := calc.Divide(0); result != 0 {
		t.Errorf("Divide(0) should return 0, got %f", result)
	}

	// Test clear
	calc.Clear()
	if mem := calc.Memory(); mem != 0 {
		t.Errorf("After Clear(), memory should be 0, got %f", mem)
	}
}

func TestCalculatorChaining(t *testing.T) {
	// Test operation chaining
	calc := NewCalculator()

	// 0 + 10 - 5 * 2 / 2 = 5
	calc.Add(10)
	calc.Subtract(5)
	calc.Multiply(2)
	result := calc.Divide(2)

	if result != 5 {
		t.Errorf("Operation chain should result in 5, got %f", result)
	}
}
