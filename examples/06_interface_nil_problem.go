package examples

import "fmt"

// MyError is a custom error type.
type MyError struct {
	Message string
}

func (e *MyError) Error() string {
	return e.Message
}

// getError demonstrates the classic nil interface trap.
// It returns a non-nil error even though the pointer is nil!
func getError(fail bool) error {
	var err *MyError = nil

	if fail {
		err = &MyError{Message: "something went wrong"}
	}

	// BUG: this returns a non-nil interface even when err is nil,
	// because the interface now has type=*MyError, value=nil.
	return err
}

// getErrorFixed shows the correct way to handle this.
func getErrorFixed(fail bool) error {
	if fail {
		return &MyError{Message: "something went wrong"}
	}
	// Return an untyped nil — this makes the interface truly nil.
	return nil
}

// RunInterfaceNilProblem demonstrates the nil interface gotcha.
func RunInterfaceNilProblem() {
	fmt.Println("=== 06: Interface Nil Problem ===")
	fmt.Println()

	// --- Demonstration 1: nil pointer inside interface ---
	fmt.Println("1. nil pointer vs nil interface:")

	var p *int = nil
	var i interface{} = p

	fmt.Printf("   p == nil:  %t\n", p == nil)       // true
	fmt.Printf("   i == nil:  %t\n", i == nil)        // false!
	fmt.Printf("   i type:    %T\n", i)               // *int
	fmt.Println("   → The interface is NOT nil because it holds type *int")

	// --- Demonstration 2: error return pattern ---
	fmt.Println("\n2. The error return trap:")

	err1 := getError(false)
	fmt.Printf("   getError(false) == nil: %t  ← BUG! Should be nil\n", err1 == nil)

	err2 := getError(true)
	fmt.Printf("   getError(true)  == nil: %t  ← correct\n", err2 == nil)

	// --- Demonstration 3: fixed version ---
	fmt.Println("\n3. Fixed version:")

	err3 := getErrorFixed(false)
	fmt.Printf("   getErrorFixed(false) == nil: %t  ← correct!\n", err3 == nil)

	err4 := getErrorFixed(true)
	fmt.Printf("   getErrorFixed(true)  == nil: %t  ← correct\n", err4 == nil)

	// --- Key lesson ---
	fmt.Println("\n   LESSON: Never return a typed nil pointer as an interface.")
	fmt.Println("   Always return a plain nil for the 'no error' case.")

	fmt.Println()
}
