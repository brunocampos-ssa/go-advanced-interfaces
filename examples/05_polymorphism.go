package examples

import (
	"fmt"
	"math"
)

// Shape defines a geometric shape that has area and perimeter.
type Shape interface {
	Area() float64
	Perimeter() float64
	Name() string
}

// Circle is a shape defined by its radius.
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }
func (c Circle) Name() string       { return "Circle" }

// Rectangle is a shape defined by width and height.
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }
func (r Rectangle) Name() string       { return "Rectangle" }

// Triangle is a shape defined by three sides.
type Triangle struct {
	A, B, C float64
}

func (t Triangle) Area() float64 {
	// Heron's formula
	s := (t.A + t.B + t.C) / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }
func (t Triangle) Name() string       { return "Triangle" }

// PrintShapeInfo prints area and perimeter for any Shape.
func PrintShapeInfo(s Shape) {
	fmt.Printf("  %-10s → Area: %8.2f | Perimeter: %8.2f\n",
		s.Name(), s.Area(), s.Perimeter())
}

// TotalArea calculates the combined area of multiple shapes.
func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// RunPolymorphism demonstrates interface-based polymorphism.
func RunPolymorphism() {
	fmt.Println("=== 05: Polymorphism ===")
	fmt.Println()

	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 3, Height: 4},
		Triangle{A: 3, B: 4, C: 5},
	}

	// The same function handles all shapes — this is polymorphism.
	for _, s := range shapes {
		PrintShapeInfo(s)
	}

	fmt.Printf("\n  Total area: %.2f\n", TotalArea(shapes))
	fmt.Println()
}
