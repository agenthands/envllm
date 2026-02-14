package ops

import (
	"encoding/json"
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
