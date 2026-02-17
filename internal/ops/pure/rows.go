package pure

import (
	"fmt"
	"github.com/agenthands/envllm/internal/runtime"
)

// SelectFields picks specific fields from a list of maps (KindRows).
func SelectFields(s *runtime.Session, source runtime.Value, fields runtime.Value) (runtime.Value, error) {
	if source.Kind != runtime.KindRows && source.Kind != runtime.KindList {
		return runtime.Value{}, fmt.Errorf("SELECT_FIELDS: source must be ROWS or LIST, got %s", source.Kind)
	}

	rawRows := source.V.([]map[string]interface{})
	
	var fieldList []string
	for _, f := range fields.V.([]runtime.Value) {
		fieldList = append(fieldList, f.V.(string))
	}

	var result []map[string]interface{}
	for _, row := range rawRows {
		newRow := make(map[string]interface{})
		for _, f := range fieldList {
			if val, ok := row[f]; ok {
				newRow[f] = val
			}
		}
		result = append(result, newRow)
	}

	return runtime.Value{Kind: runtime.KindRows, V: result}, nil
}

// FilterRows filters rows based on a simple condition.
func FilterRows(s *runtime.Session, source runtime.Value, key string, op string, val runtime.Value) (runtime.Value, error) {
	if source.Kind != runtime.KindRows {
		return runtime.Value{}, fmt.Errorf("FILTER_ROWS: source must be ROWS, got %s", source.Kind)
	}

	rows := source.V.([]map[string]interface{})
	var result []map[string]interface{}

	for _, row := range rows {
		rowVal, ok := row[key]
		if !ok {
			continue
		}

		match := false
		switch op {
		case "==":
			match = fmt.Sprintf("%v", rowVal) == fmt.Sprintf("%v", val.V)
		case "!=":
			match = fmt.Sprintf("%v", rowVal) != fmt.Sprintf("%v", val.V)
		case ">":
			// Basic numeric compare
			rv, _ := rowVal.(float64)
			vv, _ := val.V.(float64)
			match = rv > vv
		case "<":
			rv, _ := rowVal.(float64)
			vv, _ := val.V.(float64)
			match = rv < vv
		}

		if match {
			result = append(result, row)
		}
	}

	return runtime.Value{Kind: runtime.KindRows, V: result}, nil
}

// AggregateRows groups and computes simple aggregates.
func AggregateRows(s *runtime.Session, source runtime.Value, groupBy string, compute string) (runtime.Value, error) {
	if source.Kind != runtime.KindRows {
		return runtime.Value{}, fmt.Errorf("AGGREGATE_ROWS: source must be ROWS, got %s", source.Kind)
	}

	rows := source.V.([]map[string]interface{})
	groups := make(map[string][]map[string]interface{})

	for _, row := range rows {
		key := fmt.Sprintf("%v", row[groupBy])
		groups[key] = append(groups[key], row)
	}

	var result []map[string]interface{}
	for gKey, gRows := range groups {
		newRow := map[string]interface{}{
			groupBy: gKey,
		}
		
		if compute == "COUNT" {
			newRow["count"] = len(gRows)
		}
		
		result = append(result, newRow)
	}

	return runtime.Value{Kind: runtime.KindRows, V: result}, nil
}
