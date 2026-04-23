*Read this in other languages: [Portugues (BR)](README.pt-BR.md)*

# Advanced Go Internals: Interfaces and Reflection

> **Duration:** 2 hours
> **Level:** Advanced
> **Prerequisites:** Basic Go syntax, structs, methods, packages

---

## Learning Objectives

By the end of this class, students will be able to:

- Explain how Go interfaces work internally (the `(type, value)` pair)
- Distinguish between interface values and concrete values
- Use type assertions safely and recognize when they panic
- Apply type switches for runtime type branching
- Design polymorphic systems using interface-based composition
- Use the `reflect` package for struct inspection and understand its tradeoffs

---

## 1. Quick Review: Interfaces in Go

Go interfaces are **implicitly implemented**. A type satisfies an interface simply by implementing all of its methods — no `implements` keyword is needed.

This is often called **duck typing**:

> "If it walks like a duck and quacks like a duck, then it is a duck."

```go
type Speaker interface {
    Speak() string
}

type Dog struct{ Name string }

func (d Dog) Speak() string {
    return d.Name + " says: Woof!"
}

type Robot struct{ Model string }

func (r Robot) Speak() string {
    return r.Model + " says: Beep boop!"
}
```

Both `Dog` and `Robot` satisfy `Speaker` without explicitly declaring it. The compiler verifies this at assignment time.

**Key insight:** Interfaces define *behavior*, not *identity*. Any type that has the right methods qualifies.

---

## 2. How Interfaces Work Internally

Go has **two** different internal representations for interface values, defined in the runtime source code (`runtime/runtime2.go`):

- **`eface`** — for the empty interface (`interface{}` / `any`)
- **`iface`** — for interfaces with methods

### 2.1 `eface` — The Empty Interface

When you use `interface{}` or `any`, the runtime stores an **`eface`** (empty face) struct:

```
  eface (empty interface)
  ┌─────────────────────────────────────────────────────────────┐
  │                                                             │
  │   ┌──────────┐         ┌──────────────────────────────┐    │
  │   │  _type   │────────►│  runtime._type                │    │
  │   │ (*_type) │         │                               │    │
  │   └──────────┘         │  size    uintptr              │    │
  │                        │  hash    uint32               │    │
  │   ┌──────────┐         │  kind    uint8  (e.g. struct) │    │
  │   │  data    │──┐      │  str     nameOff              │    │
  │   │ (unsafe  │  │      └──────────────────────────────┘    │
  │   │ .Pointer)│  │                                          │
  │   └──────────┘  │      ┌──────────────────────────────┐    │
  │                 └─────►│  actual data                  │    │
  │                        │  e.g. int(42), "hello", etc.  │    │
  │                        └──────────────────────────────┘    │
  └─────────────────────────────────────────────────────────────┘
```

```go
// Go source:                     // Runtime representation:
var i interface{} = 42            // eface{ _type: *_type(int), data: *42 }
var s any = "hello"               // eface{ _type: *_type(string), data: *"hello" }
```

The `eface` has only **two fields** (16 bytes on 64-bit):

| Field   | Type             | Description                              |
|---------|------------------|------------------------------------------|
| `_type` | `*runtime._type` | Pointer to type metadata (size, kind, hash, etc.) |
| `data`  | `unsafe.Pointer`  | Pointer to the actual value               |

> **Note:** For small values (e.g. `int`, `bool`, pointers), Go may store the value
> directly in the `data` pointer instead of allocating heap memory — this is a
> runtime optimization called **scalar inlining**.

### 2.2 `iface` — Interfaces With Methods

When you use an interface that has methods (e.g. `Speaker`, `io.Reader`, `error`), the runtime uses an **`iface`** struct:

```
  iface (non-empty interface)
  ┌──────────────────────────────────────────────────────────────────────┐
  │                                                                      │
  │   ┌──────────┐         ┌──────────────────────────────────────┐     │
  │   │   tab    │────────►│  runtime.itab                        │     │
  │   │ (*itab)  │         │                                      │     │
  │   └──────────┘         │  inter  *interfacetype ──► Speaker   │     │
  │                        │  _type  *_type         ──► Dog       │     │
  │   ┌──────────┐         │  hash   uint32                       │     │
  │   │  data    │──┐      │  fun    [1]uintptr ──► method table  │     │
  │   │ (unsafe  │  │      │         fun[0] = Dog.Speak           │     │
  │   │ .Pointer)│  │      │         fun[1] = ...                 │     │
  │   └──────────┘  │      └──────────────────────────────────────┘     │
  │                 │                                                    │
  │                 │      ┌──────────────────────────────────────┐     │
  │                 └─────►│  actual data                         │     │
  │                        │  Dog{Name: "Rex"}                    │     │
  │                        └──────────────────────────────────────┘     │
  └──────────────────────────────────────────────────────────────────────┘
```

```go
// Go source:                     // Runtime representation:
var s Speaker = Dog{Name: "Rex"}  // iface{ tab: *itab(Speaker,Dog), data: *Dog{...} }
```

The `iface` also has **two fields** (16 bytes), but the first field points to an `itab` instead of a plain type:

| Field  | Type             | Description                                   |
|--------|------------------|-----------------------------------------------|
| `tab`  | `*runtime.itab`  | Pointer to the **interface table** (see below) |
| `data` | `unsafe.Pointer`  | Pointer to the actual value                    |

### 2.3 The `itab` — Interface Table (Virtual Method Table)

The `itab` is the key structure that enables **dynamic dispatch** in Go. It is cached per `(interface, concrete type)` pair:

```
  itab (interface table)
  ┌───────────────────────────────────────────────┐
  │  inter   *interfacetype ──► Speaker           │  which interface?
  │  _type   *_type         ──► Dog               │  which concrete type?
  │  hash    uint32          =  0x5a3c...         │  copy of _type.hash (fast switch)
  │  _       [4]byte                              │  padding
  │  fun     [n]uintptr                           │  method pointers:
  │           fun[0] ──► Dog.Speak                │    maps to Speaker.Speak()
  │           fun[1] ──► ...                      │    (one entry per interface method)
  └───────────────────────────────────────────────┘
```

When you call `s.Speak()`, the runtime:

1. Reads `s.tab` to get the `itab`
2. Looks up `tab.fun[0]` (the slot for `Speak`)
3. Calls the function pointer with `s.data` as the receiver

This is similar to a **vtable** in C++, but computed and cached at runtime.

### 2.4 `eface` vs `iface` — Side-by-Side Comparison

```
      eface (interface{} / any)              iface (Speaker, error, io.Reader, ...)
  ┌──────────────────────────┐          ┌──────────────────────────┐
  │  _type  ──► runtime._type│          │  tab  ──► runtime.itab   │
  │             (type info)  │          │           ┌─────────────┐│
  │                          │          │           │ inter (intf) ││
  │                          │          │           │ _type (conc) ││
  │                          │          │           │ fun[] (meths)││
  │                          │          │           └─────────────┘│
  ├──────────────────────────┤          ├──────────────────────────┤
  │  data  ──► actual value  │          │  data  ──► actual value  │
  └──────────────────────────┘          └──────────────────────────┘

  • No methods to dispatch               • Has method lookup table (fun[])
  • Lighter: just type + data            • Richer: type + interface + methods
  • Used by: fmt.Println, json.Marshal   • Used by: io.Reader, error, sort.Interface
```

### 2.5 Memory Layout — Complete Picture

Here is the full memory picture when assigning `Dog{Name: "Rex"}` to a `Speaker` interface:

```
  Stack                          Heap
  ┌─────────────────┐
  │ var s Speaker    │
  │ ┌─────────────┐ │           ┌─────────────────────────────┐
  │ │ tab  ───────────────────► │ itab                        │
  │ └─────────────┘ │           │   inter ──► Speaker         │
  │ ┌─────────────┐ │           │   _type ──► Dog             │
  │ │ data ──────────────┐      │   fun[0] ──► Dog.Speak      │
  │ └─────────────┘ │   │      └─────────────────────────────┘
  └─────────────────┘   │
                        │      ┌─────────────────────────────┐
                        └─────►│ Dog                         │
                               │   Name: "Rex"               │
                               └─────────────────────────────┘

  Method call: s.Speak()
  ─────────────────────
  1. Load s.tab                          → *itab
  2. Load s.tab.fun[0]                   → Dog.Speak (function pointer)
  3. Call Dog.Speak(s.data)              → "Rex says: Woof!"
```

### 2.6 Nil Interface vs Interface Holding a Nil Pointer

This is one of Go's most subtle gotchas. Understanding `eface`/`iface` makes it clear:

```
  Case 1: Nil interface               Case 2: Interface holding nil pointer
  var i interface{} = nil             var p *int = nil
                                      var j interface{} = p

  ┌──────────────────────┐            ┌──────────────────────┐
  │  _type:  nil         │            │  _type:  ──► *int    │  ← type IS set!
  │  data:   nil         │            │  data:   nil         │
  └──────────────────────┘            └──────────────────────┘
        i == nil  ✓                         j == nil  ✗ !!!

  Both fields are nil,                _type is NOT nil (it points to *int),
  so the interface IS nil.            so the interface is NOT nil,
                                      even though the pointer inside is nil.
```

```
  Case 3: The error return trap

  func getError() error {             error is an iface:
      var err *MyError = nil
      return err                      ┌──────────────────────┐
  }                                   │  tab:  ──► itab      │  ← tab IS set!
                                      │          (error,      │     (inter=error,
  err := getError()                   │           *MyError)   │      _type=*MyError)
  err == nil  →  FALSE!               │  data: nil           │
                                      └──────────────────────┘

  Fix: return nil (not a typed nil pointer)

  func getErrorFixed() error {
      return nil                      ┌──────────────────────┐
  }                                   │  tab:   nil          │  ← both nil
                                      │  data:  nil          │
  err := getErrorFixed()              └──────────────────────┘
  err == nil  →  TRUE ✓
```

**Rule:** An interface is `nil` only when **both** its internal fields (`_type`/`tab` and `data`) are `nil`.

**Practical rule:** Never return a typed nil pointer as an interface value. Always return a bare `nil`.

---

## 3. The Empty Interface

The empty interface has zero methods:

```go
interface{}
```

Since Go 1.18, there is an alias:

```go
any
```

Every type satisfies the empty interface because every type has at least zero methods.

### Use Cases

```go
func PrintAnything(v interface{}) {
    fmt.Println(v)
}

PrintAnything(42)
PrintAnything("hello")
PrintAnything([]int{1, 2, 3})
```

**Real-world examples:**
- `fmt.Println(a ...interface{})`
- `json.Marshal(v interface{}) ([]byte, error)`
- Generic data containers before Go generics

### When to Use

| Use `interface{}` / `any`           | Prefer concrete types or generics |
|--------------------------------------|-------------------------------------|
| Serialization (JSON, XML)            | Business logic                      |
| Logging frameworks                   | Domain models                       |
| Plugin systems                       | Internal APIs                       |
| Legacy code without generics         | New code (use `[T any]` instead)    |

---

## 4. Type Assertions

A type assertion extracts the concrete value from an interface:

```go
value := i.(Type)
```

### Unsafe Form (can panic)

```go
var i interface{} = "hello"

s := i.(string)  // OK: s = "hello"
n := i.(int)     // PANIC: interface holds string, not int
```

### Safe Form (never panics)

```go
s, ok := i.(string)
if ok {
    fmt.Println("It's a string:", s)
} else {
    fmt.Println("Not a string")
}
```

### When Assertions Are Useful

```go
// Check if a type implements an optional interface
type Closer interface {
    Close() error
}

func MaybeClose(v interface{}) {
    if c, ok := v.(Closer); ok {
        c.Close()
    }
}
```

**Best practice:** Always use the two-value form unless you are 100% certain of the type.

---

## 5. Type Switch

A type switch lets you branch based on the **runtime type** of an interface value:

```go
func Describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("Integer: %d\n", v)
    case string:
        fmt.Printf("String: %q\n", v)
    case bool:
        fmt.Printf("Boolean: %t\n", v)
    default:
        fmt.Printf("Unknown type: %T\n", v)
    }
}
```

### How It Differs From Regular Switch

| Regular switch           | Type switch                  |
|--------------------------|-------------------------------|
| Compares **values**      | Compares **types**            |
| `switch x { case 1: }`  | `switch v := x.(type) { }`   |
| Variable keeps its type  | Variable is narrowed to each case's type |

### Real-World Use Cases

1. **JSON parsing:** `json.Unmarshal` returns `map[string]interface{}` — type switches help walk the tree
2. **Logging systems:** Format values differently based on type
3. **Protocol handlers:** Different message types need different processing
4. **AST walkers:** Compilers and linters switch on node types

---

## 6. Polymorphism in Go

Go does not have classical OOP inheritance. Instead, it uses **interface-based polymorphism**.

### Classical OOP (NOT Go)

```
       Animal
      /      \
    Dog      Cat
```

### Go's Approach

```
  ┌──────────────┐
  │  Shape       │  (interface)
  │  Area()      │
  │  Perimeter() │
  └──────┬───────┘
         │ satisfied by
    ┌────┴────┐
    │         │
 Circle   Rectangle
```

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}

type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}
```

### Using Polymorphism

```go
func PrintShapeInfo(s Shape) {
    fmt.Printf("Area: %.2f | Perimeter: %.2f\n", s.Area(), s.Perimeter())
}

PrintShapeInfo(Circle{Radius: 5})
PrintShapeInfo(Rectangle{Width: 3, Height: 4})
```

### Composition Over Inheritance

Go encourages **small interfaces** composed together:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type ReadWriter interface {
    Reader
    Writer
}
```

**Guideline:** Keep interfaces small (1-3 methods). The standard library's `io.Reader` has just one method and is one of the most powerful abstractions in Go.

---

## 7. Reflection

The `reflect` package allows you to inspect and manipulate types and values **at runtime**.

### Core Functions

```go
import "reflect"

t := reflect.TypeOf(x)   // Returns the reflect.Type
v := reflect.ValueOf(x)  // Returns the reflect.Value
```

### Inspecting a Struct

```go
type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"email"`
    Age   int    `json:"age"`
}

u := User{Name: "Alice", Email: "alice@example.com", Age: 30}

t := reflect.TypeOf(u)
v := reflect.ValueOf(u)

fmt.Println("Type:", t.Name())           // "User"
fmt.Println("Kind:", t.Kind())           // "struct"
fmt.Println("Fields:", t.NumField())     // 3

for i := 0; i < t.NumField(); i++ {
    field := t.Field(i)
    value := v.Field(i)
    fmt.Printf("  %s (%s) = %v [tag: %s]\n",
        field.Name, field.Type, value, field.Tag)
}
```

Output:

```
Type: User
Kind: struct
Fields: 3
  Name (string) = Alice [tag: json:"name" validate:"required"]
  Email (string) = alice@example.com [tag: json:"email" validate:"email"]
  Age (int) = 30 [tag: json:"age"]
```

### Real-World Use Cases

| Use Case              | Example                          |
|------------------------|----------------------------------|
| Serialization          | `encoding/json`, `encoding/xml`  |
| ORM field mapping      | GORM, sqlx                       |
| Validation             | `go-playground/validator`        |
| Dependency injection   | Wire, dig                        |
| Testing frameworks     | testify, gomock                  |

---

## 8. When NOT to Use Reflection — Benchmarks & Tradeoffs

Reflection is powerful but comes with significant costs. This project includes a full benchmark suite in `benchmarks/` so you can measure the cost yourself.

### Running the Benchmarks

```bash
go test ./benchmarks/ -bench=. -benchmem
```

### 8.1 Benchmark Results

Results from Apple M1 Pro (your numbers will vary, but the **ratios** are consistent):

#### Group 1 — Field Access

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
DirectFieldAccess                  0.32        0       0
ReflectFieldByIndex                2.10        0       0       ← 7x slower
ReflectFieldByName                28.40        0       0       ← 90x slower
```

```
  Direct              Reflect (by index)       Reflect (by name)
  u.Name              v.Field(0).String()      v.FieldByName("Name")

  ┌──────┐            ┌──────┐                 ┌──────┐
  │ 0.3  │            │ 2.1  │  ███░░          │ 28.4 │  ██████████████████████████░░
  │ ns   │            │ ns   │                 │ ns   │
  └──────┘            └──────┘                 └──────┘
  baseline            ~7x                      ~90x
```

**Takeaway:** `FieldByName` does a string lookup on every call. If you must use reflection, prefer `Field(index)` and cache the index.

#### Group 2 — Type Identification

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
TypeAssertion                      0.32        0       0
TypeAssertionSafe (v, ok)          0.32        0       0
TypeSwitch                         0.32        0       0
ReflectValueOf                     0.75        0       0       ← 2x
ReflectTypeOf                      4.55        0       0       ← 14x
```

```
  Assertion  TypeSwitch   ValueOf       TypeOf
  i.(string) switch i.()  reflect.VOF   reflect.TOF

  ┌──────┐   ┌──────┐    ┌──────┐      ┌──────┐
  │ 0.32 │   │ 0.32 │    │ 0.75 │  ██  │ 4.55 │  ██████████████
  │ ns   │   │ ns   │    │ ns   │      │ ns   │
  └──────┘   └──────┘    └──────┘      └──────┘
  baseline   ≈ same       ~2x           ~14x
```

**Takeaway:** Type assertions and type switches compile to simple pointer comparisons — they are effectively free. Use them instead of `reflect.TypeOf` whenever possible.

#### Group 3 — Method Calls

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
DirectMethodCall                  21.19       16       1
InterfaceMethodCall               21.28       16       1       ← ~same!
ReflectMethodCall                324.30      152       6       ← 15x slower, 6 allocs
```

```
  Direct           Interface           Reflection
  d.Speak()        s.Speak()           v.MethodByName("Speak").Call(nil)

  ┌──────┐         ┌──────┐            ┌──────────┐
  │ 21.2 │  ██     │ 21.3 │  ██        │ 324.3    │  ██████████████████████████████
  │ ns   │         │ ns   │            │ ns       │
  │ 1 al │         │ 1 al │            │ 6 allocs │
  └──────┘         └──────┘            └──────────┘
  baseline         ≈ same              ~15x + 6 allocs
```

**Takeaway:** Interface dispatch through `iface.tab.fun[]` is practically free compared to a direct call (the Go compiler and CPU branch predictor handle it well). Reflection method calls are dramatically slower and allocate on every call.

#### Group 4 — Struct Iteration

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
ManualStructRead                   0.32        0       0
ReflectStructIterate              17.03        0       0       ← 53x slower
```

#### Group 5 — Cached vs Uncached reflect.TypeOf

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
ReflectTypeOfUncached              2.08        0       0
ReflectTypeOfCached                2.12        0       0       ← ≈ same
```

**Takeaway:** The runtime already caches `reflect.Type` internally, so caching it yourself in a variable doesn't improve `TypeOf` performance. However, caching **avoids repeated `TypeOf` calls** in code that also uses the result (e.g. iterating fields).

### 8.2 Performance Summary

```
  Operation                             Relative Cost
  ────────────────────────────────────────────────────
  Direct field access / type assertion  ██                          1x (baseline)
  reflect.ValueOf                       ████                        ~2x
  reflect.TypeOf                        ████████████████            ~14x
  reflect.Field(index)                  ████████                    ~7x
  reflect.FieldByName                   ████████████████████████    ~90x
  reflect.MethodByName + Call           ████████████████████████    ~15x + allocs
  ────────────────────────────────────────────────────
```

### 8.3 Complexity & Compile-Time Safety

Beyond performance, reflection hurts **readability** and **safety**:

```go
// Compile-time error (GOOD) — typo caught by compiler:
user.Naem  // ← won't compile

// Runtime panic (BAD) — no error until the code runs:
reflect.ValueOf(user).FieldByName("Naem")  // ← compiles fine, panics at runtime
```

Reflection code is:
- Hard to read and maintain
- Easy to write incorrectly
- Difficult to debug (errors surface at runtime, not compile time)

### 8.4 Best Practices

| Do                                          | Don't                                          |
|----------------------------------------------|-------------------------------------------------|
| Use reflection in libraries/frameworks       | Use reflection in application business logic    |
| Use `Field(index)` over `FieldByName`        | Use `FieldByName` in hot loops                  |
| Cache `reflect.Type` and field indices        | Call `reflect.TypeOf` repeatedly in tight loops |
| Check `Kind()` before operations              | Assume the type without checking               |
| Consider generics (Go 1.18+) as alternative  | Use reflection when concrete types work         |
| Prefer type assertions / type switches        | Use `reflect.TypeOf` just to identify a type    |

### 8.5 Decision Tree

```
Do you need runtime type inspection?
├── No → Use concrete types or generics
└── Yes
    ├── Do you know the possible types? → Use type assertion / type switch
    └── Types are unknown at compile time?
        ├── Is it a library/framework? → Reflection may be appropriate
        └── Is it application code?
            ├── Can generics solve it? → Use generics
            └── No alternative? → Use reflection with caution
                                   • Use Field(index), not FieldByName
                                   • Cache reflect.Type
                                   • Benchmark your hot path
```

---

## 9. Recap

| Concept              | Key Takeaway                                                     |
|----------------------|-------------------------------------------------------------------|
| Interfaces           | Implicitly implemented; define behavior, not identity             |
| Internal repr.       | `(type, value)` pair — nil only when both are nil                 |
| Empty interface      | `any` / `interface{}` — accepts everything, use sparingly         |
| Type assertions      | Extract concrete type; always use the `v, ok` form               |
| Type switch          | Branch on runtime type; essential for heterogeneous data          |
| Polymorphism         | Composition over inheritance; keep interfaces small               |
| Reflection           | Powerful for frameworks; avoid in business logic                  |

### Further Reading

- [The Go Blog: The Laws of Reflection](https://go.dev/blog/laws-of-reflection)
- [The Go Blog: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Go Specification: Interface Types](https://go.dev/ref/spec#Interface_types)
- [Go Proverbs](https://go-proverbs.github.io/) — "The bigger the interface, the weaker the abstraction"

---

## Exercises

The `exercises/` folder contains 3 challenges. Try to solve them before looking at `solutions/`.

| Exercise | Topic                      | File                                  |
|----------|----------------------------|---------------------------------------|
| 1        | Type Switch Logger         | `exercises/exercise1_logger.go`       |
| 2        | Notification System        | `exercises/exercise2_notifier.go`     |
| 3        | Reflection Struct Inspector| `exercises/exercise3_reflection_inspector.go` |

Run the demo to see all examples:

```bash
go run ./cmd/demo
```

---

*Happy coding!*
