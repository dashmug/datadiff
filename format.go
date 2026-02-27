package datadiff

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// rowStatus indicates the match result for a row pair.
type rowStatus int

const (
	rowMatch    rowStatus = iota // rows are equal
	rowMismatch                  // rows exist at same position but differ
	rowExtra                     // row exists in one list but not the other
)

// diffResult holds the structured outcome of comparing two datasets.
type diffResult struct {
	equal    bool
	typeName string
	columns  []string
	diffs    []rowDiff
}

// rowDiff describes the comparison outcome for one row.
type rowDiff struct {
	index    int // row index in the original list (-1 for unmatched extras)
	status   rowStatus
	valuesA  []any  // field values from listA (nil slice if row missing from A)
	valuesB  []any  // field values from listB (nil slice if row missing from B)
	mismatch []bool // per-field: true means values differ (len == len(columns))
}

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
)

// formatDiff renders a diffResult as a human-readable tabular string
// with ANSI colour highlights.
func formatDiff(result diffResult) string {
	if result.equal {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "datadiff: []%s are not equal\n\n", result.typeName)

	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	fmt.Fprint(w, " \t#\t")
	for _, column := range result.columns {
		fmt.Fprintf(w, "%s\t", column)
	}
	fmt.Fprintln(w, "")

	fmt.Fprint(w, "-\t-\t")
	for range result.columns {
		fmt.Fprint(w, "-\t")
	}
	fmt.Fprintln(w, "")

	for _, diff := range result.diffs {
		switch diff.status {
		case rowMatch:
			writeRow(w, colorize("✓", ansiGreen), fmt.Sprintf("%d", diff.index), diff.valuesA, nil, result.columns, "")
		case rowMismatch:
			writeRow(w, colorize("✗", ansiRed), fmt.Sprintf("%d", diff.index), diff.valuesA, diff.mismatch, result.columns, "← expected")
			writeRow(w, "", "", diff.valuesB, diff.mismatch, result.columns, "← actual")
		case rowExtra:
			values := diff.valuesA
			note := "← extra in expected"
			if values == nil {
				values = diff.valuesB
				note = "← extra in actual"
			}
			writeRow(w, colorize("+", ansiYellow), fmt.Sprintf("%d", diff.index), values, nil, result.columns, note)
		}
	}

	_ = w.Flush()
	return b.String()
}

func writeRow(w *tabwriter.Writer, marker, index string, values []any, mismatch []bool, columns []string, note string) {
	fmt.Fprintf(w, "%s\t%s\t", marker, index)

	for i := 0; i < len(columns); i++ {
		value := ""
		if i < len(values) {
			value = fmt.Sprintf("%v", values[i])
			if i < len(mismatch) && mismatch[i] {
				value = colorize(value, ansiRed)
			}
		}
		fmt.Fprintf(w, "%s\t", value)
	}

	if note != "" {
		fmt.Fprintf(w, "%s\t", note)
	}
	fmt.Fprintln(w, "")
}

func colorize(value, color string) string {
	return color + value + ansiReset
}
