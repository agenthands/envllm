package extension

// Manifest represents the extension's metadata and capabilities.
type Manifest struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Provides Provides          `json:"provides"`
	Compat   Compatibility     `json:"compat"`
}

// Provides defines the operations and structs offered by the extension.
type Provides struct {
	Ops     []OpDef          `json:"ops"`
	Structs []StructDef      `json:"structs"`
}

// OpDef matches the internal registry format but exported for manifest.
type OpDef struct {
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
	ResultType   string   `json:"result_type"`
	Signature    []Param  `json:"signature"`
}

type Param struct {
	Kw   string   `json:"kw"`
	Type string   `json:"type"`
	Enum []string `json:"enum,omitempty"`
}

type StructDef struct {
	Name   string            `json:"name"`
	Fields map[string]string `json:"fields"`
}

// Compatibility defines mappings for older versions.
type Compatibility struct {
	Mappings   []Mapping `json:"mappings"`
	Deprecated []Deprecation `json:"deprecated"`
}

// Mapping defines a transformation from an old version to current.
type Mapping struct {
	ID               string           `json:"id"`
	DialectRange     string           `json:"dialect_range,omitempty"`
	ExtensionRange   string           `json:"extension_range"`
	From             Pattern          `json:"from"`
	To               Target           `json:"to"`
	RewriteRuleID    string           `json:"rewrite_rule_id"`
	Notes            string           `json:"notes,omitempty"`
}

type Pattern struct {
	OpName string `json:"op_name"`
}

type Target struct {
	OpName string `json:"op_name"`
}

type Deprecation struct {
	OpName        string `json:"op_name"`
	ReplacementID string `json:"replacement_id"`
}
