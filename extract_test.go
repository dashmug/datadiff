package datadiff

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestExtract_ValidSlice(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	input := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	got, err := extract(input)
	if err != nil {
		t.Fatalf("extract returned unexpected error: %v", err)
	}

	if got.typeName != "Person" {
		t.Fatalf("typeName mismatch: got %q, want %q", got.typeName, "Person")
	}

	wantColumns := []string{"Name", "Age"}
	if !reflect.DeepEqual(got.columns, wantColumns) {
		t.Fatalf("columns mismatch: got %#v, want %#v", got.columns, wantColumns)
	}

	if len(got.rows) != 3 {
		t.Fatalf("row count mismatch: got %d, want %d", len(got.rows), 3)
	}

	wantRows := [][]any{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 35},
	}
	for i := range wantRows {
		if !reflect.DeepEqual(got.rows[i].values, wantRows[i]) {
			t.Fatalf("row %d mismatch: got %#v, want %#v", i, got.rows[i].values, wantRows[i])
		}
	}
}

func TestExtract_EmptySlice(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	var input []Person
	got, err := extract(input)
	if err != nil {
		t.Fatalf("extract returned unexpected error: %v", err)
	}

	if got.typeName != "Person" {
		t.Fatalf("typeName mismatch: got %q, want %q", got.typeName, "Person")
	}

	wantColumns := []string{"Name", "Age"}
	if !reflect.DeepEqual(got.columns, wantColumns) {
		t.Fatalf("columns mismatch: got %#v, want %#v", got.columns, wantColumns)
	}

	if len(got.rows) != 0 {
		t.Fatalf("row count mismatch: got %d, want %d", len(got.rows), 0)
	}
}

func TestExtract_NilInput(t *testing.T) {
	_, err := extract(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "nil") {
		t.Fatalf("error mismatch: got %q, want substring %q", err.Error(), "nil")
	}
}

func TestExtract_NotASlice(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{name: "string", input: "hello"},
		{name: "int", input: 42},
		{name: "struct", input: struct{ X int }{X: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extract(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), "expected slice") {
				t.Fatalf("error mismatch: got %q, want substring %q", err.Error(), "expected slice")
			}
		})
	}
}

func TestExtract_SliceOfNonStruct(t *testing.T) {
	type Person struct {
		Name string
	}

	tests := []struct {
		name  string
		input any
	}{
		{name: "slice of ints", input: []int{1, 2}},
		{name: "slice of strings", input: []string{"a", "b"}},
		{name: "slice of pointers", input: []*Person{{Name: "Alice"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extract(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), "expected slice of structs") {
				t.Fatalf("error mismatch: got %q, want substring %q", err.Error(), "expected slice of structs")
			}
		})
	}
}

func TestExtract_UnexportedFieldsSkipped(t *testing.T) {
	type Person struct {
		Name string
		age  int
		City string
	}

	input := []Person{{Name: "Alice", age: 99, City: "NYC"}}
	got, err := extract(input)
	if err != nil {
		t.Fatalf("extract returned unexpected error: %v", err)
	}

	wantColumns := []string{"Name", "City"}
	if !reflect.DeepEqual(got.columns, wantColumns) {
		t.Fatalf("columns mismatch: got %#v, want %#v", got.columns, wantColumns)
	}

	wantValues := []any{"Alice", "NYC"}
	if !reflect.DeepEqual(got.rows[0].values, wantValues) {
		t.Fatalf("row mismatch: got %#v, want %#v", got.rows[0].values, wantValues)
	}
}

func TestExtract_NoExportedFields(t *testing.T) {
	type hidden struct {
		name string
		age  int
	}

	_, err := extract([]hidden{{name: "alice", age: 30}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "has no exported fields") {
		t.Fatalf("error mismatch: got %q, want substring %q", err.Error(), "has no exported fields")
	}
}

func TestExtract_AnonymousStruct(t *testing.T) {
	input := []struct {
		X int
	}{
		{X: 10},
		{X: 20},
	}

	got, err := extract(input)
	if err != nil {
		t.Fatalf("extract returned unexpected error: %v", err)
	}

	if got.typeName != "" {
		t.Fatalf("typeName mismatch: got %q, want empty string", got.typeName)
	}

	wantColumns := []string{"X"}
	if !reflect.DeepEqual(got.columns, wantColumns) {
		t.Fatalf("columns mismatch: got %#v, want %#v", got.columns, wantColumns)
	}

	wantRows := [][]any{{10}, {20}}
	for i := range wantRows {
		if !reflect.DeepEqual(got.rows[i].values, wantRows[i]) {
			t.Fatalf("row %d mismatch: got %#v, want %#v", i, got.rows[i].values, wantRows[i])
		}
	}
}

func TestExtract_VariousFieldTypes(t *testing.T) {
	type Record struct {
		Name      string
		Count     int
		Ratio     float64
		Active    bool
		CreatedAt time.Time
	}

	now := time.Date(2026, time.February, 27, 9, 30, 0, 0, time.UTC)
	input := []Record{{
		Name:      "alpha",
		Count:     7,
		Ratio:     3.14,
		Active:    true,
		CreatedAt: now,
	}}

	got, err := extract(input)
	if err != nil {
		t.Fatalf("extract returned unexpected error: %v", err)
	}

	wantColumns := []string{"Name", "Count", "Ratio", "Active", "CreatedAt"}
	if !reflect.DeepEqual(got.columns, wantColumns) {
		t.Fatalf("columns mismatch: got %#v, want %#v", got.columns, wantColumns)
	}

	wantValues := []any{"alpha", 7, 3.14, true, now}
	if !reflect.DeepEqual(got.rows[0].values, wantValues) {
		t.Fatalf("row mismatch: got %#v, want %#v", got.rows[0].values, wantValues)
	}
}
