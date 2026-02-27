package datadiff

import "reflect"

// compare produces a diffResult from two datasets and the parsed flags.
//
// Modes:
//   - Default (strict): rows compared index-by-index; lengths must match.
//   - ignoreOrder=true: rows matched by best-fit; order does not matter.
//   - ignoreLengths=true: extra rows reported but do not set equal=false.
func compare(a, b dataset, ignoreOrder, ignoreLengths bool) diffResult {
	result := diffResult{
		equal:    true,
		typeName: a.typeName,
		columns:  a.columns,
	}

	if result.typeName == "" {
		result.typeName = b.typeName
	}
	if len(result.columns) == 0 {
		result.columns = b.columns
	}

	if ignoreOrder {
		compareUnordered(&result, a, b, ignoreLengths)
		return result
	}

	compareOrdered(&result, a, b, ignoreLengths)
	return result
}

func compareOrdered(result *diffResult, a, b dataset, ignoreLengths bool) {
	limit := len(a.rows)
	if len(b.rows) < limit {
		limit = len(b.rows)
	}

	for i := 0; i < limit; i++ {
		mismatch, mismatchCount := fieldMismatch(a.rows[i].values, b.rows[i].values, len(result.columns))
		if mismatchCount == 0 {
			result.diffs = append(result.diffs, rowDiff{
				index:   i,
				status:  rowMatch,
				valuesA: a.rows[i].values,
				valuesB: b.rows[i].values,
			})
			continue
		}

		result.equal = false
		result.diffs = append(result.diffs, rowDiff{
			index:    i,
			status:   rowMismatch,
			valuesA:  a.rows[i].values,
			valuesB:  b.rows[i].values,
			mismatch: mismatch,
		})
	}

	for i := limit; i < len(a.rows); i++ {
		result.diffs = append(result.diffs, rowDiff{
			index:   i,
			status:  rowExtra,
			valuesA: a.rows[i].values,
		})
		if !ignoreLengths {
			result.equal = false
		}
	}

	for i := limit; i < len(b.rows); i++ {
		result.diffs = append(result.diffs, rowDiff{
			index:   i,
			status:  rowExtra,
			valuesB: b.rows[i].values,
		})
		if !ignoreLengths {
			result.equal = false
		}
	}
}

func compareUnordered(result *diffResult, a, b dataset, ignoreLengths bool) {
	type unmatchedRow struct {
		index  int
		values []any
	}

	unmatched := make([]unmatchedRow, len(b.rows))
	for i := range b.rows {
		unmatched[i] = unmatchedRow{index: i, values: b.rows[i].values}
	}

	for i, rowA := range a.rows {
		if len(unmatched) == 0 {
			result.diffs = append(result.diffs, rowDiff{index: i, status: rowExtra, valuesA: rowA.values})
			if !ignoreLengths {
				result.equal = false
			}
			continue
		}

		bestCandidate := -1
		bestMismatchCount := len(result.columns) + 1
		var bestMismatch []bool
		for j, candidate := range unmatched {
			mismatch, mismatchCount := fieldMismatch(rowA.values, candidate.values, len(result.columns))
			if mismatchCount == 0 {
				bestCandidate = j
				bestMismatch = mismatch
				bestMismatchCount = 0
				break
			}

			if mismatchCount < bestMismatchCount {
				bestCandidate = j
				bestMismatch = mismatch
				bestMismatchCount = mismatchCount
			}
		}

		if bestCandidate < 0 {
			result.diffs = append(result.diffs, rowDiff{index: i, status: rowExtra, valuesA: rowA.values})
			if !ignoreLengths {
				result.equal = false
			}
			continue
		}

		candidate := unmatched[bestCandidate]
		unmatched = append(unmatched[:bestCandidate], unmatched[bestCandidate+1:]...)

		if bestMismatchCount == 0 {
			result.diffs = append(result.diffs, rowDiff{
				index:   i,
				status:  rowMatch,
				valuesA: rowA.values,
				valuesB: candidate.values,
			})
			continue
		}

		result.equal = false
		result.diffs = append(result.diffs, rowDiff{
			index:    i,
			status:   rowMismatch,
			valuesA:  rowA.values,
			valuesB:  candidate.values,
			mismatch: bestMismatch,
		})
	}

	for _, candidate := range unmatched {
		result.diffs = append(result.diffs, rowDiff{
			index:   candidate.index,
			status:  rowExtra,
			valuesB: candidate.values,
		})
		if !ignoreLengths {
			result.equal = false
		}
	}
}

func fieldMismatch(valuesA, valuesB []any, columnCount int) ([]bool, int) {
	mismatch := make([]bool, columnCount)
	mismatchCount := 0

	for i := 0; i < columnCount; i++ {
		if i >= len(valuesA) || i >= len(valuesB) {
			mismatch[i] = true
			mismatchCount++
			continue
		}

		if reflect.DeepEqual(valuesA[i], valuesB[i]) {
			continue
		}

		mismatch[i] = true
		mismatchCount++
	}

	return mismatch, mismatchCount
}
