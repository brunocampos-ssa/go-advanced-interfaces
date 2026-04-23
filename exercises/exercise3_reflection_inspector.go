package exercises

// Exercise 3 — Reflection Struct Inspector
//
// Implement the function InspectStruct that accepts any value and prints
// detailed information about its struct fields.
//
// Requirements:
//   1. Check if the value is a struct. If not, print:
//      "Error: expected a struct, got <kind>"
//
//   2. Print the struct name.
//
//   3. For each field, print:
//      - Field name
//      - Field type
//      - Field value
//      - Struct tags (if any)
//
// Expected output format for:
//
//   type Product struct {
//       Name  string  `json:"name"`
//       Price float64 `json:"price"`
//   }
//
//   InspectStruct(Product{Name: "Laptop", Price: 999.99})
//
//   Output:
//     Struct: Product
//     Field 1: Name
//       Type:  string
//       Value: Laptop
//       Tags:  json:"name"
//     Field 2: Price
//       Type:  float64
//       Value: 999.99
//       Tags:  json:"price"
//
// Hint: Use the reflect package (reflect.TypeOf, reflect.ValueOf).

func InspectStruct(v interface{}) {
	// TODO: Implement using the reflect package.
}
