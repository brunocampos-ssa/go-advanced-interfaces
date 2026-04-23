package examples

import (
	"fmt"
	"reflect"
)

// User is a sample struct with JSON and validation tags.
type User struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"email"`
	Age   int    `json:"age" validate:"min=0,max=150"`
}

// InspectType shows basic type information using reflect.
func InspectType(v interface{}) {
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	fmt.Printf("  Type:    %s\n", t.Name())
	fmt.Printf("  Kind:    %s\n", t.Kind())
	fmt.Printf("  Package: %s\n", t.PkgPath())

	if t.Kind() == reflect.Struct {
		fmt.Printf("  Fields:  %d\n", t.NumField())
		fmt.Println()

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := val.Field(i)

			fmt.Printf("    Field %d:\n", i+1)
			fmt.Printf("      Name:  %s\n", field.Name)
			fmt.Printf("      Type:  %s\n", field.Type)
			fmt.Printf("      Value: %v\n", value)

			// Read specific struct tags.
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				fmt.Printf("      JSON:  %s\n", jsonTag)
			}
			if valTag := field.Tag.Get("validate"); valTag != "" {
				fmt.Printf("      Valid: %s\n", valTag)
			}
			fmt.Println()
		}
	}
}

// RunReflection demonstrates the reflect package.
func RunReflection() {
	fmt.Println("=== 07: Reflection ===")
	fmt.Println()

	user := User{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   30,
	}

	fmt.Println("Inspecting User struct:")
	fmt.Println()
	InspectType(user)

	// Demonstrate reflect on primitive types.
	fmt.Println("Reflection on primitive types:")
	primitives := []interface{}{42, "hello", true, 3.14}
	for _, p := range primitives {
		t := reflect.TypeOf(p)
		v := reflect.ValueOf(p)
		fmt.Printf("  %v → Type: %-7s Kind: %s\n", v, t, t.Kind())
	}

	// Demonstrate reflect on a slice.
	fmt.Println("\nReflection on a slice:")
	nums := []int{10, 20, 30}
	t := reflect.TypeOf(nums)
	v := reflect.ValueOf(nums)
	fmt.Printf("  Type: %s | Kind: %s | Len: %d\n", t, t.Kind(), v.Len())
	for i := 0; i < v.Len(); i++ {
		fmt.Printf("    [%d] = %v\n", i, v.Index(i))
	}

	fmt.Println()
}
