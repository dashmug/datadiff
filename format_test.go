package datadiff

import (
	"regexp"
	"strings"
	"testing"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func TestFormatDiff_AllMatch(t *testing.T) {
	got := formatDiff(diffResult{equal: true})
	if got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestFormatDiff_SingleMismatch(t *testing.T) {
	result := diffResult{
		equal:    false,
		typeName: "Person",
		columns:  []string{"Name", "Age"},
		diffs: []rowDiff{
			{
				index:   0,
				status:  rowMatch,
				valuesA: []any{"Alice", 30},
			},
			{
				index:    1,
				status:   rowMismatch,
				valuesA:  []any{"Bob", 25},
				valuesB:  []any{"Bob", 26},
				mismatch: []bool{false, true},
			},
		},
	}

	got := formatDiff(result)
	plain := stripANSI(got)

	if !strings.Contains(plain, "✗") {
		t.Fatalf("expected mismatch marker in output, got %q", plain)
	}
	if !strings.Contains(plain, "← expected") {
		t.Fatalf("expected expected-label in output, got %q", plain)
	}
	if !strings.Contains(plain, "← actual") {
		t.Fatalf("expected actual-label in output, got %q", plain)
	}

	if !strings.Contains(got, ansiGreen+"✓"+ansiReset) {
		t.Fatalf("expected green match marker, got %q", got)
	}
	if !strings.Contains(got, ansiRed+"✗"+ansiReset) {
		t.Fatalf("expected red mismatch marker, got %q", got)
	}
	if !strings.Contains(got, ansiRed+"25"+ansiReset) {
		t.Fatalf("expected red highlighting for mismatched expected value, got %q", got)
	}
	if !strings.Contains(got, ansiRed+"26"+ansiReset) {
		t.Fatalf("expected red highlighting for mismatched actual value, got %q", got)
	}
}

func TestFormatDiff_ExtraRows(t *testing.T) {
	result := diffResult{
		equal:    false,
		typeName: "Person",
		columns:  []string{"Name", "Age"},
		diffs: []rowDiff{
			{
				index:   2,
				status:  rowExtra,
				valuesA: []any{"Eve", 40},
			},
			{
				index:   3,
				status:  rowExtra,
				valuesB: []any{"Mallory", 50},
			},
		},
	}

	got := formatDiff(result)
	plain := stripANSI(got)

	if !strings.Contains(plain, "+") {
		t.Fatalf("expected extra-row marker in output, got %q", plain)
	}
	if !strings.Contains(plain, "← extra in expected") {
		t.Fatalf("expected expected-extra label in output, got %q", plain)
	}
	if !strings.Contains(plain, "← extra in actual") {
		t.Fatalf("expected actual-extra label in output, got %q", plain)
	}

	if !strings.Contains(got, ansiYellow+"+"+ansiReset) {
		t.Fatalf("expected yellow extra marker, got %q", got)
	}
}

func TestFormatDiff_AllMismatch(t *testing.T) {
	result := diffResult{
		equal:    false,
		typeName: "Person",
		columns:  []string{"Name", "Age"},
		diffs: []rowDiff{
			{
				index:    0,
				status:   rowMismatch,
				valuesA:  []any{"Alice", 30},
				valuesB:  []any{"Alice", 31},
				mismatch: []bool{false, true},
			},
			{
				index:    1,
				status:   rowMismatch,
				valuesA:  []any{"Bob", 25},
				valuesB:  []any{"Bobby", 25},
				mismatch: []bool{true, false},
			},
		},
	}

	got := stripANSI(formatDiff(result))

	if strings.Count(got, "✗") != 2 {
		t.Fatalf("expected two mismatch rows, got output %q", got)
	}
	if strings.Count(got, "← expected") != 2 {
		t.Fatalf("expected two expected labels, got output %q", got)
	}
	if strings.Count(got, "← actual") != 2 {
		t.Fatalf("expected two actual labels, got output %q", got)
	}
}

func TestFormatDiff_HeaderFormat(t *testing.T) {
	result := diffResult{
		equal:    false,
		typeName: "Employee",
		columns:  []string{"Name"},
		diffs: []rowDiff{
			{
				index:   0,
				status:  rowMatch,
				valuesA: []any{"Alice"},
			},
		},
	}

	got := stripANSI(formatDiff(result))
	if !strings.Contains(got, "datadiff: []Employee are not equal") {
		t.Fatalf("header mismatch: got %q", got)
	}
}

func TestFormatDiff_EmptyDiff(t *testing.T) {
	result := diffResult{
		equal:    true,
		typeName: "Person",
		columns:  []string{"Name", "Age"},
		diffs:    nil,
	}

	got := formatDiff(result)
	if got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}
