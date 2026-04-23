package examples

import "fmt"

// PrintAnything accepts any value via the empty interface.
func PrintAnything(v interface{}) {
	fmt.Printf("  Value: %v (Type: %T)\n", v, v)
}

// RunEmptyInterface demonstrates the empty interface (interface{} / any).
func RunEmptyInterface() {
	fmt.Println("=== 02: Empty Interface ===")
	fmt.Println()

	// The empty interface can hold any type.
	fmt.Println("Passing different types to PrintAnything:")
	PrintAnything(42)
	PrintAnything("hello")
	PrintAnything(3.14)
	PrintAnything(true)
	PrintAnything([]int{1, 2, 3})
	PrintAnything(map[string]int{"a": 1})

	// A slice of any can hold mixed types.
	fmt.Println("\nMixed-type slice:")
	items := []any{"Go", 2024, true, 3.14}
	for _, item := range items {
		fmt.Printf("  %v (%T)\n", item, item)
	}

	fmt.Println()
}
