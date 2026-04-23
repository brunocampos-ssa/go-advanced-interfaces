*Leia em outros idiomas: [English](README.md)*

# Go Avancado: Interfaces e Reflection

> **Duracao:** 2 horas
> **Nivel:** Avancado
> **Pre-requisitos:** Sintaxe basica de Go, structs, metodos, pacotes

---

## Objetivos de Aprendizagem

Ao final desta aula, os alunos serao capazes de:

- Explicar como interfaces funcionam internamente em Go (o par `(type, value)`)
- Distinguir entre valores de interface e valores concretos
- Usar type assertions de forma segura e reconhecer quando causam panic
- Aplicar type switches para ramificacao de tipos em tempo de execucao
- Projetar sistemas polimorficos usando composicao baseada em interfaces
- Usar o pacote `reflect` para inspecao de structs e entender seus tradeoffs

---

## 1. Revisao Rapida: Interfaces em Go

Interfaces em Go sao **implementadas implicitamente**. Um tipo satisfaz uma interface simplesmente implementando todos os seus metodos — nenhuma palavra-chave `implements` e necessaria.

Isso e frequentemente chamado de **duck typing**:

> "Se anda como um pato e faz quack como um pato, entao e um pato."

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

Tanto `Dog` quanto `Robot` satisfazem `Speaker` sem declarar isso explicitamente. O compilador verifica isso no momento da atribuicao.

**Insight principal:** Interfaces definem *comportamento*, nao *identidade*. Qualquer tipo que tenha os metodos corretos se qualifica.

---

## 2. Como Interfaces Funcionam Internamente

Go possui **duas** representacoes internas diferentes para valores de interface, definidas no codigo-fonte do runtime (`runtime/runtime2.go`):

- **`eface`** — para a interface vazia (`interface{}` / `any`)
- **`iface`** — para interfaces com metodos

### 2.1 `eface` — A Interface Vazia

Quando voce usa `interface{}` ou `any`, o runtime armazena uma struct **`eface`** (empty face):

```
  eface (interface vazia)
  ┌─────────────────────────────────────────────────────────────┐
  │                                                             │
  │   ┌──────────┐         ┌──────────────────────────────┐    │
  │   │  _type   │────────►│  runtime._type                │    │
  │   │ (*_type) │         │                               │    │
  │   └──────────┘         │  size    uintptr              │    │
  │                        │  hash    uint32               │    │
  │   ┌──────────┐         │  kind    uint8  (ex: struct)  │    │
  │   │  data    │──┐      │  str     nameOff              │    │
  │   │ (unsafe  │  │      └──────────────────────────────┘    │
  │   │ .Pointer)│  │                                          │
  │   └──────────┘  │      ┌──────────────────────────────┐    │
  │                 └─────►│  dados reais                  │    │
  │                        │  ex: int(42), "hello", etc.   │    │
  │                        └──────────────────────────────┘    │
  └─────────────────────────────────────────────────────────────┘
```

```go
// Codigo Go:                        // Representacao no runtime:
var i interface{} = 42               // eface{ _type: *_type(int), data: *42 }
var s any = "hello"                  // eface{ _type: *_type(string), data: *"hello" }
```

A `eface` possui apenas **dois campos** (16 bytes em 64-bit):

| Campo   | Tipo             | Descricao                                        |
|---------|------------------|--------------------------------------------------|
| `_type` | `*runtime._type` | Ponteiro para metadados do tipo (size, kind, hash, etc.) |
| `data`  | `unsafe.Pointer`  | Ponteiro para o valor real                        |

> **Nota:** Para valores pequenos (ex: `int`, `bool`, ponteiros), Go pode armazenar o valor
> diretamente no ponteiro `data` ao inves de alocar memoria no heap — isso e uma
> otimizacao do runtime chamada **scalar inlining**.

### 2.2 `iface` — Interfaces Com Metodos

Quando voce usa uma interface que possui metodos (ex: `Speaker`, `io.Reader`, `error`), o runtime usa uma struct **`iface`**:

```
  iface (interface com metodos)
  ┌──────────────────────────────────────────────────────────────────────┐
  │                                                                      │
  │   ┌──────────┐         ┌──────────────────────────────────────┐     │
  │   │   tab    │────────►│  runtime.itab                        │     │
  │   │ (*itab)  │         │                                      │     │
  │   └──────────┘         │  inter  *interfacetype ──► Speaker   │     │
  │                        │  _type  *_type         ──► Dog       │     │
  │   ┌──────────┐         │  hash   uint32                       │     │
  │   │  data    │──┐      │  fun    [1]uintptr ──► tabela metodos│     │
  │   │ (unsafe  │  │      │         fun[0] = Dog.Speak           │     │
  │   │ .Pointer)│  │      │         fun[1] = ...                 │     │
  │   └──────────┘  │      └──────────────────────────────────────┘     │
  │                 │                                                    │
  │                 │      ┌──────────────────────────────────────┐     │
  │                 └─────►│  dados reais                         │     │
  │                        │  Dog{Name: "Rex"}                    │     │
  │                        └──────────────────────────────────────┘     │
  └──────────────────────────────────────────────────────────────────────┘
```

```go
// Codigo Go:                            // Representacao no runtime:
var s Speaker = Dog{Name: "Rex"}         // iface{ tab: *itab(Speaker,Dog), data: *Dog{...} }
```

A `iface` tambem possui **dois campos** (16 bytes), mas o primeiro campo aponta para uma `itab` ao inves de um tipo simples:

| Campo  | Tipo             | Descricao                                        |
|--------|------------------|--------------------------------------------------|
| `tab`  | `*runtime.itab`  | Ponteiro para a **tabela de interface** (veja abaixo) |
| `data` | `unsafe.Pointer`  | Ponteiro para o valor real                        |

### 2.3 A `itab` — Tabela de Interface (Tabela de Metodos Virtuais)

A `itab` e a estrutura-chave que habilita o **dispatch dinamico** em Go. Ela e cacheada por par `(interface, tipo concreto)`:

```
  itab (tabela de interface)
  ┌───────────────────────────────────────────────┐
  │  inter   *interfacetype ──► Speaker           │  qual interface?
  │  _type   *_type         ──► Dog               │  qual tipo concreto?
  │  hash    uint32          =  0x5a3c...         │  copia de _type.hash (switch rapido)
  │  _       [4]byte                              │  padding
  │  fun     [n]uintptr                           │  ponteiros de metodos:
  │           fun[0] ──► Dog.Speak                │    mapeia para Speaker.Speak()
  │           fun[1] ──► ...                      │    (uma entrada por metodo da interface)
  └───────────────────────────────────────────────┘
```

Quando voce chama `s.Speak()`, o runtime:

1. Le `s.tab` para obter a `itab`
2. Consulta `tab.fun[0]` (o slot para `Speak`)
3. Chama o ponteiro de funcao com `s.data` como receiver

Isso e similar a uma **vtable** em C++, mas computada e cacheada em tempo de execucao.

### 2.4 `eface` vs `iface` — Comparacao Lado a Lado

```
      eface (interface{} / any)              iface (Speaker, error, io.Reader, ...)
  ┌──────────────────────────┐          ┌──────────────────────────┐
  │  _type  ──► runtime._type│          │  tab  ──► runtime.itab   │
  │             (info tipo)  │          │           ┌─────────────┐│
  │                          │          │           │ inter (intf) ││
  │                          │          │           │ _type (conc) ││
  │                          │          │           │ fun[] (mets) ││
  │                          │          │           └─────────────┘│
  ├──────────────────────────┤          ├──────────────────────────┤
  │  data  ──► valor real    │          │  data  ──► valor real    │
  └──────────────────────────┘          └──────────────────────────┘

  • Sem metodos para despachar            • Possui tabela de metodos (fun[])
  • Mais leve: apenas tipo + dados        • Mais rica: tipo + interface + metodos
  • Usada por: fmt.Println, json.Marshal  • Usada por: io.Reader, error, sort.Interface
```

### 2.5 Layout de Memoria — Visao Completa

Aqui esta a visao completa da memoria ao atribuir `Dog{Name: "Rex"}` a uma interface `Speaker`:

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

  Chamada de metodo: s.Speak()
  ────────────────────────────
  1. Carrega s.tab                        → *itab
  2. Carrega s.tab.fun[0]                 → Dog.Speak (ponteiro de funcao)
  3. Chama Dog.Speak(s.data)              → "Rex says: Woof!"
```

### 2.6 Interface Nil vs Interface Contendo Ponteiro Nil

Esta e uma das armadilhas mais sutis do Go. Entender `eface`/`iface` torna tudo claro:

```
  Caso 1: Interface nil                Caso 2: Interface contendo ponteiro nil
  var i interface{} = nil             var p *int = nil
                                      var j interface{} = p

  ┌──────────────────────┐            ┌──────────────────────┐
  │  _type:  nil         │            │  _type:  ──► *int    │  ← tipo ESTA definido!
  │  data:   nil         │            │  data:   nil         │
  └──────────────────────┘            └──────────────────────┘
        i == nil  ✓                         j == nil  ✗ !!!

  Ambos os campos sao nil,            _type NAO e nil (aponta para *int),
  entao a interface E nil.            entao a interface NAO e nil,
                                      mesmo que o ponteiro interno seja nil.
```

```
  Caso 3: A armadilha do retorno de error

  func getError() error {             error e uma iface:
      var err *MyError = nil
      return err                      ┌──────────────────────┐
  }                                   │  tab:  ──► itab      │  ← tab ESTA definido!
                                      │          (error,      │     (inter=error,
  err := getError()                   │           *MyError)   │      _type=*MyError)
  err == nil  →  FALSE!               │  data: nil           │
                                      └──────────────────────┘

  Correcao: retorne nil (nao um ponteiro nil tipado)

  func getErrorFixed() error {
      return nil                      ┌──────────────────────┐
  }                                   │  tab:   nil          │  ← ambos nil
                                      │  data:  nil          │
  err := getErrorFixed()              └──────────────────────┘
  err == nil  →  TRUE ✓
```

**Regra:** Uma interface e `nil` somente quando **ambos** os seus campos internos (`_type`/`tab` e `data`) sao `nil`.

**Regra pratica:** Nunca retorne um ponteiro nil tipado como valor de interface. Sempre retorne um `nil` puro.

---

## 3. A Interface Vazia

A interface vazia nao possui nenhum metodo:

```go
interface{}
```

Desde o Go 1.18, existe um alias:

```go
any
```

Todo tipo satisfaz a interface vazia porque todo tipo possui pelo menos zero metodos.

### Casos de Uso

```go
func PrintAnything(v interface{}) {
    fmt.Println(v)
}

PrintAnything(42)
PrintAnything("hello")
PrintAnything([]int{1, 2, 3})
```

**Exemplos do mundo real:**
- `fmt.Println(a ...interface{})`
- `json.Marshal(v interface{}) ([]byte, error)`
- Containers de dados genericos antes dos generics do Go

### Quando Usar

| Use `interface{}` / `any`            | Prefira tipos concretos ou generics |
|---------------------------------------|---------------------------------------|
| Serializacao (JSON, XML)              | Logica de negocios                    |
| Frameworks de logging                 | Modelos de dominio                    |
| Sistemas de plugins                   | APIs internas                         |
| Codigo legado sem generics            | Codigo novo (use `[T any]`)           |

---

## 4. Type Assertions (Assercoes de Tipo)

Uma type assertion extrai o valor concreto de uma interface:

```go
value := i.(Type)
```

### Forma Insegura (pode causar panic)

```go
var i interface{} = "hello"

s := i.(string)  // OK: s = "hello"
n := i.(int)     // PANIC: a interface contem string, nao int
```

### Forma Segura (nunca causa panic)

```go
s, ok := i.(string)
if ok {
    fmt.Println("E uma string:", s)
} else {
    fmt.Println("Nao e uma string")
}
```

### Quando Assercoes Sao Uteis

```go
// Verificar se um tipo implementa uma interface opcional
type Closer interface {
    Close() error
}

func MaybeClose(v interface{}) {
    if c, ok := v.(Closer); ok {
        c.Close()
    }
}
```

**Boa pratica:** Sempre use a forma com dois valores, a menos que voce tenha 100% de certeza do tipo.

---

## 5. Type Switch (Switch de Tipo)

Um type switch permite ramificar a logica com base no **tipo em tempo de execucao** de um valor de interface:

```go
func Describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("Inteiro: %d\n", v)
    case string:
        fmt.Printf("String: %q\n", v)
    case bool:
        fmt.Printf("Booleano: %t\n", v)
    default:
        fmt.Printf("Tipo desconhecido: %T\n", v)
    }
}
```

### Diferenca do Switch Regular

| Switch regular           | Type switch                    |
|--------------------------|--------------------------------|
| Compara **valores**      | Compara **tipos**              |
| `switch x { case 1: }`  | `switch v := x.(type) { }`    |
| Variavel mantem seu tipo | Variavel e estreitada para o tipo de cada case |

### Casos de Uso no Mundo Real

1. **Parsing de JSON:** `json.Unmarshal` retorna `map[string]interface{}` — type switches ajudam a percorrer a arvore
2. **Sistemas de logging:** Formatam valores de forma diferente baseado no tipo
3. **Handlers de protocolo:** Diferentes tipos de mensagem precisam de processamento diferente
4. **AST walkers:** Compiladores e linters fazem switch nos tipos de nos

---

## 6. Polimorfismo em Go

Go nao possui heranca classica de POO. Em vez disso, usa **polimorfismo baseado em interfaces**.

### POO Classica (NAO e Go)

```
       Animal
      /      \
    Dog      Cat
```

### Abordagem do Go

```
  ┌──────────────┐
  │  Shape       │  (interface)
  │  Area()      │
  │  Perimeter() │
  └──────┬───────┘
         │ satisfeita por
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

### Usando Polimorfismo

```go
func PrintShapeInfo(s Shape) {
    fmt.Printf("Area: %.2f | Perimetro: %.2f\n", s.Area(), s.Perimeter())
}

PrintShapeInfo(Circle{Radius: 5})
PrintShapeInfo(Rectangle{Width: 3, Height: 4})
```

### Composicao Sobre Heranca

Go encoraja **interfaces pequenas** compostas entre si:

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

**Diretriz:** Mantenha interfaces pequenas (1-3 metodos). O `io.Reader` da biblioteca padrao possui apenas um metodo e e uma das abstracoes mais poderosas do Go.

---

## 7. Reflection (Reflexao)

O pacote `reflect` permite inspecionar e manipular tipos e valores **em tempo de execucao**.

### Funcoes Principais

```go
import "reflect"

t := reflect.TypeOf(x)   // Retorna o reflect.Type
v := reflect.ValueOf(x)  // Retorna o reflect.Value
```

### Inspecionando uma Struct

```go
type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"email"`
    Age   int    `json:"age"`
}

u := User{Name: "Alice", Email: "alice@example.com", Age: 30}

t := reflect.TypeOf(u)
v := reflect.ValueOf(u)

fmt.Println("Tipo:", t.Name())            // "User"
fmt.Println("Kind:", t.Kind())            // "struct"
fmt.Println("Campos:", t.NumField())      // 3

for i := 0; i < t.NumField(); i++ {
    field := t.Field(i)
    value := v.Field(i)
    fmt.Printf("  %s (%s) = %v [tag: %s]\n",
        field.Name, field.Type, value, field.Tag)
}
```

Saida:

```
Tipo: User
Kind: struct
Campos: 3
  Name (string) = Alice [tag: json:"name" validate:"required"]
  Email (string) = alice@example.com [tag: json:"email" validate:"email"]
  Age (int) = 30 [tag: json:"age"]
```

### Casos de Uso no Mundo Real

| Caso de Uso            | Exemplo                          |
|-------------------------|----------------------------------|
| Serializacao            | `encoding/json`, `encoding/xml`  |
| Mapeamento de campos ORM| GORM, sqlx                       |
| Validacao               | `go-playground/validator`        |
| Injecao de dependencia  | Wire, dig                        |
| Frameworks de teste     | testify, gomock                  |

---

## 8. Quando NAO Usar Reflection — Benchmarks e Tradeoffs

Reflection e poderoso, mas vem com custos significativos. Este projeto inclui uma suite completa de benchmarks em `benchmarks/` para que voce possa medir o custo por conta propria.

### Executando os Benchmarks

```bash
go test ./benchmarks/ -bench=. -benchmem
```

### 8.1 Resultados dos Benchmarks

Resultados de um Apple M1 Pro (seus numeros podem variar, mas as **proporcoes** sao consistentes):

#### Grupo 1 — Acesso a Campos

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
DirectFieldAccess                  0.32        0       0
ReflectFieldByIndex                2.10        0       0       ← 7x mais lento
ReflectFieldByName                28.40        0       0       ← 90x mais lento
```

```
  Direto              Reflect (por indice)     Reflect (por nome)
  u.Name              v.Field(0).String()      v.FieldByName("Name")

  ┌──────┐            ┌──────┐                 ┌──────┐
  │ 0.3  │            │ 2.1  │  ███░░          │ 28.4 │  ██████████████████████████░░
  │ ns   │            │ ns   │                 │ ns   │
  └──────┘            └──────┘                 └──────┘
  baseline            ~7x                      ~90x
```

**Conclusao:** `FieldByName` faz uma busca por string a cada chamada. Se precisar usar reflection, prefira `Field(index)` e faca cache do indice.

#### Grupo 2 — Identificacao de Tipo

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
  baseline   ≈ igual      ~2x           ~14x
```

**Conclusao:** Type assertions e type switches compilam para simples comparacoes de ponteiros — sao efetivamente gratuitos. Use-os ao inves de `reflect.TypeOf` sempre que possivel.

#### Grupo 3 — Chamadas de Metodo

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
DirectMethodCall                  21.19       16       1
InterfaceMethodCall               21.28       16       1       ← ~igual!
ReflectMethodCall                324.30      152       6       ← 15x mais lento, 6 alocacoes
```

```
  Direto           Interface           Reflection
  d.Speak()        s.Speak()           v.MethodByName("Speak").Call(nil)

  ┌──────┐         ┌──────┐            ┌──────────┐
  │ 21.2 │  ██     │ 21.3 │  ██        │ 324.3    │  ██████████████████████████████
  │ ns   │         │ ns   │            │ ns       │
  │ 1 al │         │ 1 al │            │ 6 alocs  │
  └──────┘         └──────┘            └──────────┘
  baseline         ≈ igual              ~15x + 6 alocacoes
```

**Conclusao:** O dispatch de interface atraves de `iface.tab.fun[]` e praticamente gratuito comparado a uma chamada direta (o compilador Go e o branch predictor da CPU lidam bem com isso). Chamadas de metodo via reflection sao dramaticamente mais lentas e alocam memoria a cada chamada.

#### Grupo 4 — Iteracao de Struct

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
ManualStructRead                   0.32        0       0
ReflectStructIterate              17.03        0       0       ← 53x mais lento
```

#### Grupo 5 — reflect.TypeOf com Cache vs sem Cache

```
Benchmark                         ns/op     B/op   allocs/op
─────────────────────────────────────────────────────────────
ReflectTypeOfUncached              2.08        0       0
ReflectTypeOfCached                2.12        0       0       ← ≈ igual
```

**Conclusao:** O runtime ja faz cache de `reflect.Type` internamente, entao armazena-lo voce mesmo em uma variavel nao melhora a performance do `TypeOf`. Porem, fazer cache **evita chamadas repetidas de `TypeOf`** em codigo que tambem usa o resultado (ex: iterar campos).

### 8.2 Resumo de Performance

```
  Operacao                                Custo Relativo
  ────────────────────────────────────────────────────────
  Acesso direto / type assertion          ██                          1x (baseline)
  reflect.ValueOf                         ████                        ~2x
  reflect.TypeOf                          ████████████████            ~14x
  reflect.Field(index)                    ████████                    ~7x
  reflect.FieldByName                     ████████████████████████    ~90x
  reflect.MethodByName + Call             ████████████████████████    ~15x + alocacoes
  ────────────────────────────────────────────────────────
```

### 8.3 Complexidade e Seguranca em Tempo de Compilacao

Alem da performance, reflection prejudica **legibilidade** e **seguranca**:

```go
// Erro em tempo de compilacao (BOM) — typo capturado pelo compilador:
user.Naem  // ← nao compila

// Panic em tempo de execucao (RUIM) — sem erro ate o codigo rodar:
reflect.ValueOf(user).FieldByName("Naem")  // ← compila normalmente, panic em runtime
```

Codigo com reflection e:
- Dificil de ler e manter
- Facil de escrever incorretamente
- Dificil de debugar (erros aparecem em runtime, nao em tempo de compilacao)

### 8.4 Boas Praticas

| Faca                                         | Nao Faca                                       |
|----------------------------------------------|-------------------------------------------------|
| Use reflection em bibliotecas/frameworks     | Use reflection em logica de negocios            |
| Use `Field(index)` ao inves de `FieldByName` | Use `FieldByName` em loops criticos             |
| Faca cache de `reflect.Type` e indices        | Chame `reflect.TypeOf` repetidamente em loops   |
| Verifique `Kind()` antes das operacoes        | Assuma o tipo sem verificar                     |
| Considere generics (Go 1.18+) como alternativa| Use reflection quando tipos concretos funcionam |
| Prefira type assertions / type switches       | Use `reflect.TypeOf` apenas para identificar tipo|

### 8.5 Arvore de Decisao

```
Voce precisa de inspecao de tipo em runtime?
├── Nao → Use tipos concretos ou generics
└── Sim
    ├── Voce conhece os tipos possiveis? → Use type assertion / type switch
    └── Tipos desconhecidos em tempo de compilacao?
        ├── E uma biblioteca/framework? → Reflection pode ser apropriado
        └── E codigo de aplicacao?
            ├── Generics resolvem? → Use generics
            └── Sem alternativa? → Use reflection com cautela
                                   • Use Field(index), nao FieldByName
                                   • Faca cache de reflect.Type
                                   • Faca benchmark do seu hot path
```

---

## 9. Recapitulacao

| Conceito             | Ponto Principal                                                   |
|----------------------|-------------------------------------------------------------------|
| Interfaces           | Implementadas implicitamente; definem comportamento, nao identidade|
| Repr. interna        | Par `(type, value)` — nil somente quando ambos sao nil             |
| Interface vazia      | `any` / `interface{}` — aceita tudo, use com moderacao             |
| Type assertions      | Extraem tipo concreto; sempre use a forma `v, ok`                  |
| Type switch          | Ramifica por tipo em runtime; essencial para dados heterogeneos    |
| Polimorfismo         | Composicao sobre heranca; mantenha interfaces pequenas             |
| Reflection           | Poderoso para frameworks; evite em logica de negocios              |

### Leitura Complementar

- [The Go Blog: The Laws of Reflection](https://go.dev/blog/laws-of-reflection)
- [The Go Blog: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Go Specification: Interface Types](https://go.dev/ref/spec#Interface_types)
- [Go Proverbs](https://go-proverbs.github.io/) — "The bigger the interface, the weaker the abstraction" (Quanto maior a interface, mais fraca a abstracao)

---

## Exercicios

A pasta `exercises/` contem 3 desafios. Tente resolve-los antes de olhar as `solutions/`.

| Exercicio | Topico                        | Arquivo                                 |
|-----------|-------------------------------|-----------------------------------------|
| 1         | Logger com Type Switch        | `exercises/exercise1_logger.go`         |
| 2         | Sistema de Notificacoes       | `exercises/exercise2_notifier.go`       |
| 3         | Inspetor de Struct com Reflect| `exercises/exercise3_reflection_inspector.go` |

Execute o demo para ver todos os exemplos:

```bash
go run ./cmd/demo
```

---

*Bom codigo!*
