package main

import (
	"fmt"
	"github.com/brunocampos-ssa/go-advanced-interfaces/examples"
	"strings"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Advanced Go Internals: Interfaces and Reflection")
	fmt.Println("  Demo Program")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	examples.RunBasicInterface()
	examples.RunEmptyInterface()
	examples.RunTypeAssertion()
	examples.RunTypeSwitch()
	examples.RunPolymorphism()
	examples.RunInterfaceNilProblem()
	examples.RunReflection()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Demo complete! Now try the exercises in exercises/")
	fmt.Println(strings.Repeat("=", 60))
}
