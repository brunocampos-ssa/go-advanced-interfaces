package examples

import "fmt"

// Describe uses a type switch to print information based on the runtime type.
func Describe(i interface{}) {
	switch v := i.(type) {
	case int:
		fmt.Printf("  Integer: %d (doubled: %d)\n", v, v*2)
	case string:
		fmt.Printf("  String: %q (length: %d)\n", v, len(v))
	case bool:
		fmt.Printf("  Boolean: %t\n", v)
	case []int:
		fmt.Printf("  Int slice: %v (length: %d)\n", v, len(v))
	case nil:
		fmt.Println("  Nil value")
	default:
		fmt.Printf("  Unknown type: %T = %v\n", v, v)
	}
}

// RunTypeSwitch demonstrates type switches with various types.
func RunTypeSwitch() {
	fmt.Println("=== 04: Type Switch ===")
	fmt.Println()

	values := []interface{}{
		42,
		"hello",
		true,
		[]int{1, 2, 3},
		nil,
		3.14,
		struct{ Name string }{"anonymous"},
	}

	for _, v := range values {
		Describe(v)
	}

	fmt.Println()
}
