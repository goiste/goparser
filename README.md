# Go Parser

A small package that helps you parse Go files by annotation in doc comment, e.g. get var values or list of struct
methods.

<br>

Features:

- get values of variables:
    - literal types
    - slices of literal types
    - maps with literal types as keys and values
- get list of function names:
    - by method receiver type
    - by parameters types

<br>

Literal types:

- bool
- string
- int8-int64
- uint8-uint64
- float32/64

  — set directly, w/o type castings or pointers

<br>

Usage:

parsed code:

```go
var (
// parser
boolValue = true

// parser:str
stringValue = "3"
```

[full example_code.go](example/example_code.go)

main code:

```go
import (
	gp "github.com/goiste/goparser"
)

...

p, err := gp.New("example_code.go")
if err != nil {
panic(err)
}

boolValues := gp.GetBasicValues[bool](p, "parser")
for _, v := range boolValues {
fmt.Printf("name: %s; value: %t\n", v.Name, v.Value) // name: boolValue; value: true
}

stringValues := gp.GetBasicValues[string](p, "parser", "parser:str")
for _, v := range stringValues {
fmt.Printf("name: %s; value: %q\n", v.Name, v.Value) // name: stringValue; value: "3"
}
```

[full example.go](example/example.go)