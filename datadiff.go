// Package datadiff provides test assertions for comparing slices of structs,
// with rich tabular diff output that highlights row and column mismatches.
//
// Inspired by github.com/MrPowers/chispa for Python DataFrames.
package datadiff

import "testing"

// Version is the current module version.
const Version = "0.1.0-dev"

// Flag controls comparison behavior in [Assert].
type Flag int

const (
	// IgnoreOrder compares lists without regard to element position.
	// Rows are matched by finding the closest counterpart in the other list.
	IgnoreOrder Flag = iota + 1

	// IgnoreLengths allows lists of different lengths.
	// Extra rows are reported in the diff output but do not cause the
	// assertion to fail. Without this flag, differing lengths are a failure.
	IgnoreLengths
)

// Assert compares listA and listB and reports differences through t.
// Both arguments must be slices of the same struct type.
//
// By default, comparison is strict: rows must appear in the same order
// and both lists must have equal length. Pass [IgnoreOrder] and/or
// [IgnoreLengths] to relax those constraints.
//
// Assert calls t.Fatalf for programmer errors (invalid flags or invalid
// inputs) and t.Errorf for data mismatches.
//
// Returns true if the lists are equal under the given flags, false otherwise.
func Assert(t *testing.T, listA, listB any, flags ...any) bool {
	t.Helper()

	var ignoreOrder, ignoreLengths bool
	for _, f := range flags {
		flag, ok := f.(Flag)
		if !ok {
			t.Fatalf("datadiff: unknown flag type %T (expected datadiff.Flag)", f)
			return false
		}

		switch flag {
		case IgnoreOrder:
			ignoreOrder = true
		case IgnoreLengths:
			ignoreLengths = true
		default:
			t.Fatalf("datadiff: unknown flag value: %d", flag)
			return false
		}
	}

	dsA, err := extract(listA)
	if err != nil {
		t.Fatalf("datadiff: first argument: %v", err)
		return false
	}

	dsB, err := extract(listB)
	if err != nil {
		t.Fatalf("datadiff: second argument: %v", err)
		return false
	}

	if dsA.typeName != dsB.typeName {
		t.Errorf("datadiff: type mismatch: []%s vs []%s", dsA.typeName, dsB.typeName)
		return false
	}

	result := compare(dsA, dsB, ignoreOrder, ignoreLengths)
	if !result.equal {
		t.Errorf("\n%s", formatDiff(result))
		return false
	}

	return true
}
