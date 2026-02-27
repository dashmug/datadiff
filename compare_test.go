package datadiff

import (
	"reflect"
	"testing"
)

func makeDataset(typeName string, columns []string, rows ...[]any) dataset {
	ds := dataset{
		typeName: typeName,
		columns:  columns,
		rows:     make([]row, len(rows)),
	}

	for i := range rows {
		ds.rows[i] = row{values: rows[i]}
	}

	return ds
}

func makePersonDataset(rows ...[]any) dataset {
	return makeDataset("Person", []string{"Name", "Age"}, rows...)
}

func TestCompare_StrictEqual(t *testing.T) {
	a := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
	)
	b := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
	)

	got := compare(a, b, false, false)
	if !got.equal {
		t.Fatal("expected equal=true")
	}
	if len(got.diffs) != 2 {
		t.Fatalf("diff count mismatch: got %d, want %d", len(got.diffs), 2)
	}

	for i, diff := range got.diffs {
		if diff.status != rowMatch {
			t.Fatalf("row %d status mismatch: got %v, want %v", i, diff.status, rowMatch)
		}
		if diff.index != i {
			t.Fatalf("row %d index mismatch: got %d, want %d", i, diff.index, i)
		}
	}
}

func TestCompare_StrictMismatch(t *testing.T) {
	a := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
	)
	b := makePersonDataset(
		[]any{"Alice", 31},
		[]any{"Bob", 25},
	)

	got := compare(a, b, false, false)
	if got.equal {
		t.Fatal("expected equal=false")
	}
	if got.diffs[0].status != rowMismatch {
		t.Fatalf("expected first row mismatch, got %v", got.diffs[0].status)
	}

	wantMismatch := []bool{false, true}
	if !reflect.DeepEqual(got.diffs[0].mismatch, wantMismatch) {
		t.Fatalf("mismatch flags mismatch: got %#v, want %#v", got.diffs[0].mismatch, wantMismatch)
	}
	if got.diffs[1].status != rowMatch {
		t.Fatalf("expected second row match, got %v", got.diffs[1].status)
	}
}

func TestCompare_StrictDifferentLengths(t *testing.T) {
	a := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
		[]any{"Charlie", 35},
	)
	b := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
	)

	got := compare(a, b, false, false)
	if got.equal {
		t.Fatal("expected equal=false")
	}
	if len(got.diffs) != 3 {
		t.Fatalf("diff count mismatch: got %d, want %d", len(got.diffs), 3)
	}

	extra := got.diffs[2]
	if extra.status != rowExtra {
		t.Fatalf("expected extra row status, got %v", extra.status)
	}
	if extra.valuesA == nil || extra.valuesB != nil {
		t.Fatalf("expected extra-in-expected row, got valuesA=%#v valuesB=%#v", extra.valuesA, extra.valuesB)
	}
}

func TestCompare_StrictDifferentLengths_IgnoreLengths(t *testing.T) {
	a := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
		[]any{"Charlie", 35},
	)
	b := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
	)

	got := compare(a, b, false, true)
	if !got.equal {
		t.Fatal("expected equal=true when ignoreLengths=true and overlap matches")
	}
	if len(got.diffs) != 3 {
		t.Fatalf("diff count mismatch: got %d, want %d", len(got.diffs), 3)
	}
	if got.diffs[2].status != rowExtra {
		t.Fatalf("expected extra row status, got %v", got.diffs[2].status)
	}
}

func TestCompare_StrictBothEmpty(t *testing.T) {
	a := makePersonDataset()
	b := makePersonDataset()

	got := compare(a, b, false, false)
	if !got.equal {
		t.Fatal("expected equal=true")
	}
	if len(got.diffs) != 0 {
		t.Fatalf("expected no diffs, got %d", len(got.diffs))
	}
}

func TestCompare_StrictOneEmpty(t *testing.T) {
	a := makePersonDataset()
	b := makePersonDataset([]any{"Alice", 30})

	got := compare(a, b, false, false)
	if got.equal {
		t.Fatal("expected equal=false")
	}
	if len(got.diffs) != 1 {
		t.Fatalf("diff count mismatch: got %d, want %d", len(got.diffs), 1)
	}
	if got.diffs[0].status != rowExtra {
		t.Fatalf("expected rowExtra status, got %v", got.diffs[0].status)
	}
	if got.diffs[0].valuesA != nil || got.diffs[0].valuesB == nil {
		t.Fatalf("expected extra-in-actual row, got valuesA=%#v valuesB=%#v", got.diffs[0].valuesA, got.diffs[0].valuesB)
	}
}

func TestCompare_UnorderedExactMatch(t *testing.T) {
	a := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
		[]any{"Charlie", 35},
	)
	b := makePersonDataset(
		[]any{"Charlie", 35},
		[]any{"Alice", 30},
		[]any{"Bob", 25},
	)

	got := compare(a, b, true, false)
	if !got.equal {
		t.Fatal("expected equal=true")
	}
	if len(got.diffs) != 3 {
		t.Fatalf("diff count mismatch: got %d, want %d", len(got.diffs), 3)
	}
	for i, diff := range got.diffs {
		if diff.status != rowMatch {
			t.Fatalf("row %d status mismatch: got %v, want %v", i, diff.status, rowMatch)
		}
	}
}

func TestCompare_UnorderedMismatch(t *testing.T) {
	a := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
	)
	b := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Charlie", 25},
	)

	got := compare(a, b, true, false)
	if got.equal {
		t.Fatal("expected equal=false")
	}
	if len(got.diffs) != 2 {
		t.Fatalf("diff count mismatch: got %d, want %d", len(got.diffs), 2)
	}
	if got.diffs[1].status != rowMismatch {
		t.Fatalf("expected second row mismatch, got %v", got.diffs[1].status)
	}
}

func TestCompare_UnorderedDuplicates(t *testing.T) {
	a := makePersonDataset(
		[]any{"X", 1},
		[]any{"X", 1},
		[]any{"Y", 2},
	)
	b := makePersonDataset(
		[]any{"Y", 2},
		[]any{"X", 1},
		[]any{"X", 1},
	)

	got := compare(a, b, true, false)
	if !got.equal {
		t.Fatal("expected equal=true for duplicated unordered rows")
	}

	for i, diff := range got.diffs {
		if diff.status != rowMatch {
			t.Fatalf("row %d status mismatch: got %v, want %v", i, diff.status, rowMatch)
		}
	}
}

func TestCompare_UnorderedPartialMatch(t *testing.T) {
	a := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
		[]any{"Charlie", 35},
	)
	b := makePersonDataset(
		[]any{"Bob", 25},
		[]any{"Alice", 30},
		[]any{"Chuck", 35},
	)

	got := compare(a, b, true, false)
	if got.equal {
		t.Fatal("expected equal=false")
	}

	matchCount := 0
	mismatchCount := 0
	for _, diff := range got.diffs {
		switch diff.status {
		case rowMatch:
			matchCount++
		case rowMismatch:
			mismatchCount++
		}
	}

	if matchCount != 2 || mismatchCount != 1 {
		t.Fatalf("status counts mismatch: matches=%d mismatches=%d", matchCount, mismatchCount)
	}
}

func TestCompare_UnorderedDifferentLengths(t *testing.T) {
	a := makePersonDataset(
		[]any{"Alice", 30},
		[]any{"Bob", 25},
	)
	b := makePersonDataset(
		[]any{"Bob", 25},
		[]any{"Alice", 30},
		[]any{"Charlie", 35},
	)

	got := compare(a, b, true, true)
	if !got.equal {
		t.Fatal("expected equal=true with ignoreOrder+ignoreLengths when common rows match")
	}

	extraCount := 0
	for _, diff := range got.diffs {
		if diff.status == rowExtra {
			extraCount++
		}
	}
	if extraCount != 1 {
		t.Fatalf("extra row count mismatch: got %d, want %d", extraCount, 1)
	}
}

func TestCompare_MismatchFieldTracking(t *testing.T) {
	a := makeDataset("Person", []string{"Name", "Age", "City"}, []any{"Alice", 30, "NY"})
	b := makeDataset("Person", []string{"Name", "Age", "City"}, []any{"Alicia", 30, "SF"})

	got := compare(a, b, false, false)
	if got.equal {
		t.Fatal("expected equal=false")
	}
	if len(got.diffs) != 1 || got.diffs[0].status != rowMismatch {
		t.Fatalf("expected one mismatch row, got %#v", got.diffs)
	}

	wantMismatch := []bool{true, false, true}
	if !reflect.DeepEqual(got.diffs[0].mismatch, wantMismatch) {
		t.Fatalf("mismatch tracking mismatch: got %#v, want %#v", got.diffs[0].mismatch, wantMismatch)
	}
}
