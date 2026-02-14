package runtime

import (
	"encoding/json"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestExecResult_Schema(t *testing.T) {
	schemaPath := "../../schemas/exec_result.schema.json"
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to compile schema: %v", err)
	}

	res := ExecResult{
		SchemaVersion: "obs-0.1",
		Status:        "ok",
		VarsDelta:     make(map[string]Value),
		Budgets:       make(map[string]BudgetStats),
		Events:        []Event{},
		Errors:        []Error{},
		Truncated:     TruncationFlags{},
	}

	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if err := schema.Validate(v); err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}
