package ops

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/agenthands/rlm-go/internal/runtime"
)

// Table holds all registered operations and their signatures.
type Table struct {
	Version string         `json:"version"`
	Ops     map[string]*Op `json:"-"`
}

// Op represents a single operation definition.
type Op struct {
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
	ResultType   runtime.Kind  `json:"result_type"`
	Signature    []Param  `json:"signature"`
	Into         bool     `json:"into"`
}

// Param represents a keyword-type pair in an operation signature.
type Param struct {
	Kw   string       `json:"kw"`
	Type runtime.Kind `json:"type,omitempty"`
	Enum []string     `json:"enum,omitempty"`
}

// LoadTable reads and parses the ops.json file.
func LoadTable(path string) (*Table, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Version string `json:"version"`
		Ops     []Op   `json:"ops"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	t := &Table{
		Version: raw.Version,
		Ops:     make(map[string]*Op),
	}

	for _, op := range raw.Ops {
		o := op
		t.Ops[op.Name] = &o
	}

	return t, nil
}

// ValidatedKwArg represents a keyword-value pair that has been type-checked.
type ValidatedKwArg struct {
	Keyword string
	Value   runtime.Value
}

// ValidateSignature checks if an operation statement matches its definition.
func (t *Table) ValidateSignature(name string, args []ValidatedKwArg) (*Op, error) {
	op, ok := t.Ops[name]
	if !ok {
		return nil, fmt.Errorf("unknown operation: %s", name)
	}

	if len(args) != len(op.Signature) {
		return nil, fmt.Errorf("%s: expected %d arguments, got %d", name, len(op.Signature), len(args))
	}

	for i, param := range op.Signature {
		arg := args[i]
		if arg.Keyword != param.Kw {
			return nil, fmt.Errorf("%s: argument %d keyword mismatch: expected %s, got %s", name, i, param.Kw, arg.Keyword)
		}

		// Type checking
		if param.Type != "" && arg.Value.Kind != param.Type {
			return nil, fmt.Errorf("%s: argument %s type mismatch: expected %s, got %s", name, param.Kw, param.Type, arg.Value.Kind)
		}

		// Enum checking
		if len(param.Enum) > 0 {
			val, ok := arg.Value.V.(string)
			if !ok {
				return nil, fmt.Errorf("%s: argument %s must be a string for enum check", name, param.Kw)
			}
			found := false
			for _, e := range param.Enum {
				if e == val {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("%s: argument %s invalid enum value: %s", name, param.Kw, val)
			}
		}
	}

	return op, nil
}
