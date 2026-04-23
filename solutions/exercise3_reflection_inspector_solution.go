package solutions

import (
	"fmt"
	"reflect"
)

// InspectStruct prints detailed information about a struct's fields.
func InspectStruct(v interface{}) {
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	// Handle pointer to struct.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}

	if t.Kind() != reflect.Struct {
		fmt.Printf("Error: expected a struct, got %s\n", t.Kind())
		return
	}

	fmt.Printf("Struct: %s\n", t.Name())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := val.Field(i)

		fmt.Printf("Field %d: %s\n", i+1, field.Name)
		fmt.Printf("  Type:  %s\n", field.Type)
		fmt.Printf("  Value: %v\n", value)

		if field.Tag != "" {
			fmt.Printf("  Tags:  %s\n", field.Tag)
		}
	}
}

// Product is a sample struct for testing the inspector.
type Product struct {
	Name  string  `json:"name" validate:"required"`
	Price float64 `json:"price" validate:"min=0"`
	Stock int     `json:"stock"`
}

// RunInspectorSolution demonstrates the solution for Exercise 3.
func RunInspectorSolution() {
	fmt.Println("=== Exercise 3 Solution: Reflection Inspector ===")
	fmt.Println()

	product := Product{
		Name:  "Laptop",
		Price: 999.99,
		Stock: 42,
	}

	InspectStruct(product)

	fmt.Println("\nInspecting a non-struct:")
	InspectStruct("not a struct")

	fmt.Println()
}
