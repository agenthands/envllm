package runtime

import (
	"encoding/json"
)

// ExecResult represents the result of executing an RLM program or cell.
type ExecResult struct {
	SchemaVersion string                 `json:"schema_version"`
	Cell          CellInfo               `json:"cell"`
	Status        string                 `json:"status"`
	VarsDelta     map[string]Value       `json:"vars_delta"`
	Result        *Value                 `json:"result,omitempty"`
	Final         *Value                 `json:"final,omitempty"`
	Budgets       map[string]BudgetStats `json:"budgets"`
	Events        []Event                `json:"events"`
	Errors        []Error                `json:"errors"`
	Truncated     TruncationFlags        `json:"truncated"`
}

type CellInfo struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

type BudgetStats struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

type Event struct {
	T      string `json:"t"` // "op", "subcall", etc.
	Op     string `json:"op,omitempty"`
	Into   string `json:"into,omitempty"`
	MS     int    `json:"ms,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Loc     struct {
		File string `json:"file"`
		Line int    `json:"line"`
		Col  int    `json:"col"`
	} `json:"loc,omitempty"`
	Hint string `json:"hint,omitempty"`
}

type TruncationFlags struct {
	Obs      bool `json:"obs"`
	Prints   bool `json:"prints"`
	Previews bool `json:"previews"`
}

func (r ExecResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
