package benchmarks

import (
	"fmt"
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// Shared types used across benchmarks
// ---------------------------------------------------------------------------

// Speaker is a simple interface with one method.
type Speaker interface {
	Speak() string
}

// Dog satisfies Speaker.
type Dog struct {
	Name string
}

func (d Dog) Speak() string {
	return d.Name + " says: Woof!"
}

// User is a struct used for field-access benchmarks.
type User struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"email"`
	Age   int    `json:"age"`
}

// ---------------------------------------------------------------------------
// 1. Direct Field Access vs Reflection Field Access
// ---------------------------------------------------------------------------

// BenchmarkDirectFieldAccess reads a struct field directly.
func BenchmarkDirectFieldAccess(b *testing.B) {
	u := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	var s string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = u.Name
	}
	_ = s
}

// BenchmarkReflectFieldByIndex reads a struct field via reflect using a
// numeric index (the fastest reflection path).
func BenchmarkReflectFieldByIndex(b *testing.B) {
	u := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	v := reflect.ValueOf(u)
	var s string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = v.Field(0).String()
	}
	_ = s
}

// BenchmarkReflectFieldByName reads a struct field via reflect.FieldByName
// (slower — requires a map lookup by name on every call).
func BenchmarkReflectFieldByName(b *testing.B) {
	u := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	v := reflect.ValueOf(u)
	var s string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = v.FieldByName("Name").String()
	}
	_ = s
}

// ---------------------------------------------------------------------------
// 2. Type Assertion vs Type Switch vs Reflection TypeOf
// ---------------------------------------------------------------------------

// BenchmarkTypeAssertion extracts a concrete type from an interface using
// a direct type assertion.
func BenchmarkTypeAssertion(b *testing.B) {
	var i interface{} = "hello"
	var s string

	b.ResetTimer()
	for i2 := 0; i2 < b.N; i2++ {
		s = i.(string)
	}
	_ = s
}

// BenchmarkTypeAssertionSafe uses the two-value (safe) form of type assertion.
func BenchmarkTypeAssertionSafe(b *testing.B) {
	var i interface{} = "hello"
	var s string

	b.ResetTimer()
	for i2 := 0; i2 < b.N; i2++ {
		s, _ = i.(string)
	}
	_ = s
}

// BenchmarkTypeSwitch uses a type switch to identify the type.
func BenchmarkTypeSwitch(b *testing.B) {
	var i interface{} = "hello"
	var s string

	b.ResetTimer()
	for i2 := 0; i2 < b.N; i2++ {
		switch v := i.(type) {
		case string:
			s = v
		case int:
			s = fmt.Sprintf("%d", v)
		case bool:
			s = fmt.Sprintf("%t", v)
		default:
			s = "unknown"
		}
	}
	_ = s
}

// BenchmarkReflectTypeOf uses reflect.TypeOf to determine the type at runtime.
func BenchmarkReflectTypeOf(b *testing.B) {
	var i interface{} = "hello"
	var name string

	b.ResetTimer()
	for i2 := 0; i2 < b.N; i2++ {
		name = reflect.TypeOf(i).String()
	}
	_ = name
}

// BenchmarkReflectValueOf wraps a value with reflect.ValueOf and reads it back.
func BenchmarkReflectValueOf(b *testing.B) {
	var i interface{} = "hello"
	var s string

	b.ResetTimer()
	for i2 := 0; i2 < b.N; i2++ {
		s = reflect.ValueOf(i).String()
	}
	_ = s
}

// ---------------------------------------------------------------------------
// 3. Interface Method Call vs Direct Method Call
// ---------------------------------------------------------------------------

// BenchmarkDirectMethodCall calls a method on a concrete type (no dispatch).
func BenchmarkDirectMethodCall(b *testing.B) {
	d := Dog{Name: "Rex"}
	var s string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = d.Speak()
	}
	_ = s
}

// BenchmarkInterfaceMethodCall calls a method through an interface (iface dispatch).
func BenchmarkInterfaceMethodCall(b *testing.B) {
	var s Speaker = Dog{Name: "Rex"}
	var result string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = s.Speak()
	}
	_ = result
}

// BenchmarkReflectMethodCall calls a method using reflect.Value.MethodByName.
func BenchmarkReflectMethodCall(b *testing.B) {
	var s Speaker = Dog{Name: "Rex"}
	v := reflect.ValueOf(s)
	var result string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = v.MethodByName("Speak").Call(nil)[0].String()
	}
	_ = result
}

// ---------------------------------------------------------------------------
// 4. Struct Iteration: Manual vs Reflection
// ---------------------------------------------------------------------------

// BenchmarkManualStructRead reads all fields of a struct manually.
func BenchmarkManualStructRead(b *testing.B) {
	u := User{Name: "Alice", Email: "alice@example.com", Age: 30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = u.Name
		_ = u.Email
		_ = u.Age
	}
}

// BenchmarkReflectStructIterate reads all fields using reflection.
func BenchmarkReflectStructIterate(b *testing.B) {
	u := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	v := reflect.ValueOf(u)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < v.NumField(); j++ {
			_ = v.Field(j).Interface()
		}
	}
}

// ---------------------------------------------------------------------------
// 5. Cached Reflection vs Uncached Reflection
// ---------------------------------------------------------------------------

// BenchmarkReflectTypeOfUncached calls reflect.TypeOf on every iteration.
func BenchmarkReflectTypeOfUncached(b *testing.B) {
	u := User{Name: "Alice", Email: "alice@example.com", Age: 30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := reflect.TypeOf(u)
		_ = t.NumField()
	}
}

// BenchmarkReflectTypeOfCached calls reflect.TypeOf once and reuses it.
func BenchmarkReflectTypeOfCached(b *testing.B) {
	u := User{Name: "Alice", Email: "alice@example.com", Age: 30}
	t := reflect.TypeOf(u)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.NumField()
	}
}
