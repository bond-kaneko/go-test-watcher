package main

// Calculator represents a simple calculator with memory
type Calculator struct {
	memory float64
}

// NewCalculator creates a new calculator with zero memory
func NewCalculator() *Calculator {
	return &Calculator{memory: 0}
}

// Add adds a number to memory and returns the result
func (c *Calculator) Add(value float64) float64 {
	c.memory += value
	return c.memory
}

// Subtract subtracts a number from memory and returns the result
func (c *Calculator) Subtract(value float64) float64 {
	c.memory -= value
	return c.memory
}

// Multiply multiplies memory by a number and returns the result
func (c *Calculator) Multiply(value float64) float64 {
	c.memory *= value
	return c.memory
}

// Divide divides memory by a number and returns the result
// Returns an error (zero) if dividing by zero
func (c *Calculator) Divide(value float64) float64 {
	if value == 0 {
		return 0
	}
	c.memory /= value
	return c.memory
}

// Clear resets memory to zero
func (c *Calculator) Clear() {
	c.memory = 0
}

// Memory returns the current value in memory
func (c *Calculator) Memory() float64 {
	return c.memory
}
