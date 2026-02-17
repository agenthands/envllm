package validate

import (
	"encoding/json"
	"fmt"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Validator struct {
	schema *jsonschema.Schema
}

func NewValidator(schemaPath string) (*Validator, error) {
	compiler := jsonschema.NewCompiler()
	
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return nil, err
	}

	return &Validator{schema: schema}, nil
}

func (v *Validator) ValidateAST(prog *ast.Program) error {
	cellNames := make(map[string]bool)
	if prog.Task != nil {
		for _, item := range prog.Task.Body {
			if cell, ok := item.(*ast.Cell); ok {
				if cellNames[cell.Name] {
					return fmt.Errorf("%s: duplicate cell name: %s", cell.Loc, cell.Name)
				}
				cellNames[cell.Name] = true
			}
		}
	}

	data, err := json.Marshal(prog)
	if err != nil {
		return err
	}

	var x interface{}
	_ = json.Unmarshal(data, &x) // This is safe as we just marshaled it

	return v.schema.Validate(x)
}
