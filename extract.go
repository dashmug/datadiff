package datadiff

import (
	"fmt"
	"reflect"
)

// dataset represents a normalized list of structs for comparison.
type dataset struct {
	typeName string   // struct type name, e.g. "Person"
	columns  []string // exported field names in declaration order
	rows     []row
}

// row holds the field values for a single struct element.
type row struct {
	values []any // one value per column, same order as dataset.columns
}

// extract validates that v is a slice of structs and returns a dataset.
//
// Errors:
//   - v is nil
//   - v is not a slice
//   - slice element type is not a struct (pointers to structs are not accepted)
//   - struct has zero exported fields
func extract(v any) (dataset, error) {
	if v == nil {
		return dataset{}, fmt.Errorf("datadiff: input is nil")
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Slice {
		return dataset{}, fmt.Errorf("datadiff: expected slice, got %T", v)
	}

	elemType := reflect.TypeOf(v).Elem()
	if elemType.Kind() != reflect.Struct {
		return dataset{}, fmt.Errorf("datadiff: expected slice of structs, got slice of %s", elemType.Kind())
	}

	columns := make([]string, 0, elemType.NumField())
	exportedFieldIndexes := make([]int, 0, elemType.NumField())
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		if !field.IsExported() {
			continue
		}
		exportedFieldIndexes = append(exportedFieldIndexes, i)
		columns = append(columns, field.Name)
	}

	if len(columns) == 0 {
		return dataset{}, fmt.Errorf("datadiff: struct %s has no exported fields", elemType.Name())
	}

	result := dataset{
		typeName: elemType.Name(),
		columns:  columns,
		rows:     make([]row, value.Len()),
	}

	for i := 0; i < value.Len(); i++ {
		element := value.Index(i)
		values := make([]any, len(exportedFieldIndexes))
		for j, fieldIndex := range exportedFieldIndexes {
			values[j] = element.Field(fieldIndex).Interface()
		}
		result.rows[i] = row{values: values}
	}

	return result, nil
}
