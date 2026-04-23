package examples

import "fmt"

// RunTypeAssertion demonstrates safe and unsafe type assertions.
func RunTypeAssertion() {
	fmt.Println("=== 03: Type Assertion ===")
	fmt.Println()

	var i interface{} = "hello, Go!"

	// --- Safe assertion (two-value form) ---
	fmt.Println("Safe assertion (string):")
	s, ok := i.(string)
	if ok {
		fmt.Printf("  Success: %q\n", s)
	}

	fmt.Println("Safe assertion (int):")
	n, ok := i.(int)
	if ok {
		fmt.Printf("  Success: %d\n", n)
	} else {
		fmt.Println("  Failed: value is not an int")
	}

	// --- Unsafe assertion (single-value form) ---
	fmt.Println("\nUnsafe assertion (string) — this works:")
	s2 := i.(string)
	fmt.Printf("  Got: %q\n", s2)

	fmt.Println("\nUnsafe assertion (int) — this would PANIC:")
	fmt.Println("  // n2 := i.(int)  ← uncommenting this causes a runtime panic")

	// --- Practical use: checking optional interface ---
	fmt.Println("\nPractical: checking if a type implements an optional interface")

	var w interface{} = Dog{Name: "Buddy"}

	if speaker, ok := w.(Speaker); ok {
		fmt.Printf("  It speaks: %s\n", speaker.Speak())
	} else {
		fmt.Println("  Does not implement Speaker")
	}

	fmt.Println()
}
