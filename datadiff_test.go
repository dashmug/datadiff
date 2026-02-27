package datadiff

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

type Person struct {
	Name string
	Age  int
}

type Employee struct {
	Name string
	Age  int
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestAssert_IdenticalSlices(t *testing.T) {
	a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
	b := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}

	if !Assert(t, a, b) {
		t.Fatal("expected Assert to return true for identical slices")
	}
}

func TestAssert_EmptySlices(t *testing.T) {
	a := []Person{}
	b := []Person{}

	if !Assert(t, a, b) {
		t.Fatal("expected Assert to return true for empty slices")
	}
}

func TestAssert_DifferentValues(t *testing.T) {
	assertScenarioFails(t, "different-values", "datadiff: []Person are not equal", "← expected", "← actual")
}

func TestAssert_DifferentLengths(t *testing.T) {
	assertScenarioFails(t, "different-lengths", "datadiff: []Person are not equal", "← extra in expected")
}

func TestAssert_DifferentOrder(t *testing.T) {
	assertScenarioFails(t, "different-order", "datadiff: []Person are not equal")
}

func TestAssert_IgnoreOrder_SameElements(t *testing.T) {
	a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}, {Name: "Charlie", Age: 35}}
	b := []Person{{Name: "Charlie", Age: 35}, {Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}

	if !Assert(t, a, b, IgnoreOrder) {
		t.Fatal("expected Assert to return true with IgnoreOrder for same elements")
	}
}

func TestAssert_IgnoreOrder_MissingElement(t *testing.T) {
	assertScenarioFails(t, "ignore-order-missing", "datadiff: []Person are not equal")
}

func TestAssert_IgnoreLengths_ExtraInA(t *testing.T) {
	a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}, {Name: "Charlie", Age: 35}}
	b := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}

	if !Assert(t, a, b, IgnoreLengths) {
		t.Fatal("expected Assert to return true with IgnoreLengths when overlap matches")
	}
}

func TestAssert_IgnoreLengths_MismatchInOverlap(t *testing.T) {
	assertScenarioFails(t, "ignore-lengths-mismatch", "datadiff: []Person are not equal")
}

func TestAssert_IgnoreOrderAndLengths(t *testing.T) {
	a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}, {Name: "Charlie", Age: 35}}
	b := []Person{{Name: "Bob", Age: 25}, {Name: "Alice", Age: 30}}

	if !Assert(t, a, b, IgnoreOrder, IgnoreLengths) {
		t.Fatal("expected Assert to return true with IgnoreOrder+IgnoreLengths when common rows match")
	}
}

func TestAssert_NilInput(t *testing.T) {
	assertScenarioFails(t, "nil-input", "datadiff: first argument: datadiff: input is nil")
}

func TestAssert_NotASlice(t *testing.T) {
	assertScenarioFails(t, "not-a-slice", "datadiff: first argument: datadiff: expected slice")
}

func TestAssert_TypeMismatch(t *testing.T) {
	assertScenarioFails(t, "type-mismatch", "datadiff: type mismatch: []Person vs []Employee")
}

func TestAssert_InvalidFlag(t *testing.T) {
	assertScenarioFails(t, "invalid-flag", "datadiff: unknown flag type string (expected datadiff.Flag)")
}

func TestAssert_SliceOfNonStruct(t *testing.T) {
	assertScenarioFails(t, "slice-of-non-struct", "datadiff: first argument: datadiff: expected slice of structs")
}

func TestAssert_OutputContainsTable(t *testing.T) {
	output := assertScenarioFails(t, "output-table", "datadiff: []Person are not equal")
	if !strings.Contains(output, "Name") || !strings.Contains(output, "Age") {
		t.Fatalf("expected output table columns in failure output, got: %s", output)
	}
	if !strings.Contains(output, "Alice") || !strings.Contains(output, "Bob") {
		t.Fatalf("expected row values in failure output, got: %s", output)
	}
	if !strings.Contains(output, "← expected") || !strings.Contains(output, "← actual") {
		t.Fatalf("expected expected/actual labels in failure output, got: %s", output)
	}
}

func assertScenarioFails(t *testing.T, scenario string, requiredSubstrings ...string) string {
	t.Helper()

	output, err := runAssertSubprocess(t, scenario)
	if err == nil {
		t.Fatalf("expected scenario %q to fail, got success with output: %s", scenario, output)
	}

	for _, required := range requiredSubstrings {
		if !strings.Contains(output, required) {
			t.Fatalf("expected output for scenario %q to contain %q, got: %s", scenario, required, output)
		}
	}

	return output
}

func runAssertSubprocess(t *testing.T, scenario string) (string, error) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^TestAssert_SubprocessHarness$")
	cmd.Env = append(os.Environ(),
		"DATADIFF_ASSERT_SUBPROCESS=1",
		"DATADIFF_ASSERT_SCENARIO="+scenario,
	)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestAssert_SubprocessHarness(t *testing.T) {
	if os.Getenv("DATADIFF_ASSERT_SUBPROCESS") != "1" {
		t.Skip("subprocess harness")
	}

	scenario := os.Getenv("DATADIFF_ASSERT_SCENARIO")
	switch scenario {
	case "different-values", "output-table":
		a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
		b := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 26}}
		if Assert(t, a, b) {
			t.Fatal("expected Assert to return false")
		}
	case "different-lengths":
		a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}, {Name: "Charlie", Age: 35}}
		b := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
		if Assert(t, a, b) {
			t.Fatal("expected Assert to return false")
		}
	case "different-order":
		a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
		b := []Person{{Name: "Bob", Age: 25}, {Name: "Alice", Age: 30}}
		if Assert(t, a, b) {
			t.Fatal("expected Assert to return false")
		}
	case "ignore-order-missing":
		a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
		b := []Person{{Name: "Alice", Age: 30}, {Name: "Charlie", Age: 25}}
		if Assert(t, a, b, IgnoreOrder) {
			t.Fatal("expected Assert to return false")
		}
	case "ignore-lengths-mismatch":
		a := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 26}, {Name: "Charlie", Age: 35}}
		b := []Person{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
		if Assert(t, a, b, IgnoreLengths) {
			t.Fatal("expected Assert to return false")
		}
	case "nil-input":
		Assert(t, nil, []Person{})
		t.Fatal("expected Assert to fatal for nil input")
	case "not-a-slice":
		Assert(t, Person{Name: "Alice", Age: 30}, []Person{})
		t.Fatal("expected Assert to fatal for non-slice input")
	case "type-mismatch":
		a := []Person{{Name: "Alice", Age: 30}}
		b := []Employee{{Name: "Alice", Age: 30}}
		if Assert(t, a, b) {
			t.Fatal("expected Assert to return false")
		}
	case "invalid-flag":
		Assert(t, []Person{}, []Person{}, "bad-flag")
		t.Fatal("expected Assert to fatal for invalid flag type")
	case "slice-of-non-struct":
		Assert(t, []int{1, 2}, []int{1, 2})
		t.Fatal("expected Assert to fatal for non-struct slice")
	default:
		t.Fatalf("unknown subprocess scenario %q", scenario)
	}
}
