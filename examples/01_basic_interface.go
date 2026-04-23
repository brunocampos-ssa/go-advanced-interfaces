package examples

import "fmt"

// Speaker defines a behavior: anything that can speak.
type Speaker interface {
	Speak() string
}

// Dog is a concrete type that satisfies Speaker.
type Dog struct {
	Name string
}

func (d Dog) Speak() string {
	return d.Name + " says: Woof!"
}

// Robot is another concrete type that satisfies Speaker.
type Robot struct {
	Model string
}

func (r Robot) Speak() string {
	return r.Model + " says: Beep boop!"
}

// greet accepts any Speaker — both Dog and Robot qualify.
func greet(s Speaker) {
	fmt.Println(s.Speak())
}

// RunBasicInterface demonstrates implicit interface implementation.
func RunBasicInterface() {
	fmt.Println("=== 01: Basic Interface ===")
	fmt.Println()

	dog := Dog{Name: "Rex"}
	robot := Robot{Model: "R2-D2"}

	// Both satisfy Speaker without an explicit "implements" declaration.
	greet(dog)
	greet(robot)

	// We can also store them in a slice of Speaker.
	speakers := []Speaker{dog, robot}
	fmt.Println("\nAll speakers:")
	for _, s := range speakers {
		fmt.Println(" -", s.Speak())
	}

	fmt.Println()
}
