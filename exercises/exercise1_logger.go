package exercises

// Exercise 1 — Generic Logger
//
// Implement the function LogValue that accepts any type and prints
// a description based on the runtime type.
//
// Rules:
//   - If the value is an int    → print: "[LOG] number: <value>"
//   - If the value is a string  → print: "[LOG] text: <value>"
//   - If the value is a bool   → print: "[LOG] boolean: <value>"
//   - For any other type        → print: "[LOG] unknown: <value>"
//
// You MUST use a type switch.
//
// Example:
//   LogValue(42)      → [LOG] number: 42
//   LogValue("hello") → [LOG] text: hello
//   LogValue(true)    → [LOG] boolean: true
//   LogValue(3.14)    → [LOG] unknown: 3.14

func LogValue(v interface{}) {
	// TODO: Implement using a type switch.
}
