# datadiff

[![License:
MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go
Reference](https://pkg.go.dev/badge/github.com/dashmug/datadiff.svg)](https://pkg.go.dev/github.com/dashmug/datadiff)
[![CI](https://github.com/dashmug/datadiff/actions/workflows/ci.yml/badge.svg)](https://github.com/dashmug/datadiff/actions/workflows/ci.yml)

A Go test assertion library for comparing lists of structs with rich
tabular diff output.

## Install

```bash
go get github.com/dashmug/datadiff
```

## Usage

```go
package yourpkg

import (
	"testing"

	"github.com/dashmug/datadiff"
)

type Person struct {
	Name string
	Age  int
	City string
}

func TestPeople(t *testing.T) {
	expected := []Person{
		{Name: "Alice", Age: 30, City: "New York"},
		{Name: "Bob", Age: 25, City: "Boston"},
	}

	actual := []Person{
		{Name: "Alice", Age: 30, City: "New York"},
		{Name: "Bob", Age: 26, City: "Boston"},
	}

	if !datadiff.Assert(t, expected, actual) {
		return
	}
}
```

When the assertion fails, output is tabular and highlights mismatched
rows and columns:

```text
datadiff: []Person are not equal

   #  Name   Age  City
-  -  -      -    -
✓  0  Alice  30   New York
✗  1  Bob    25   Boston  ← expected
      Bob    26   Boston  ← actual
```

## Flags

`Assert` is strict by default: order and length must match.

### IgnoreOrder

Use `IgnoreOrder` to compare rows regardless of position.

```go
ok := datadiff.Assert(t, expected, actual, datadiff.IgnoreOrder)
```

### IgnoreLengths

Use `IgnoreLengths` to allow extras while still comparing overlapping
rows.

```go
ok := datadiff.Assert(t, expected, actual, datadiff.IgnoreLengths)
```

You can combine both flags:

```go
ok := datadiff.Assert(t, expected, actual, datadiff.IgnoreOrder, datadiff.IgnoreLengths)
```

## Inspiration

This project is inspired by
[chispa](https://github.com/MrPowers/chispa), a Python DataFrame
assertion helper.

## License

MIT
