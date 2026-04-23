package solutions

import "fmt"

// LogValue prints a log entry based on the runtime type of v.
func LogValue(v interface{}) {
	switch val := v.(type) {
	case int:
		fmt.Printf("[LOG] number: %d\n", val)
	case string:
		fmt.Printf("[LOG] text: %s\n", val)
	case bool:
		fmt.Printf("[LOG] boolean: %t\n", val)
	default:
		fmt.Printf("[LOG] unknown: %v\n", val)
	}
}

// RunLoggerSolution demonstrates the solution for Exercise 1.
func RunLoggerSolution() {
	fmt.Println("=== Exercise 1 Solution: Generic Logger ===")
	fmt.Println()

	LogValue(42)
	LogValue("hello")
	LogValue(true)
	LogValue(3.14)
	LogValue([]int{1, 2, 3})

	fmt.Println()
}
