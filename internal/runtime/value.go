package runtime

import (
	"encoding/json"
	"fmt"
)

// Kind represents the type of an RLM value.
type Kind string

const (
	KindInt    Kind = "INT"
	KindBool   Kind = "BOOL"
	KindText   Kind = "TEXT"
	KindJSON   Kind = "JSON"
	KindSpan   Kind = "SPAN"
	KindString Kind = "STRING"
	KindList   Kind = "LIST"
	KindNull   Kind = "NULL"
	KindOffset Kind = "OFFSET"
)

// Value represents a typed value in the RLM runtime.
type Value struct {
	Kind Kind
	V    interface{}
}

// Span represents a range in text.
type Span struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// TextHandle represents a reference to text in the TextStore.
type TextHandle struct {
	ID           string `json:"id"`
	Bytes        int    `json:"bytes"`
	Preview      string `json:"preview,omitempty"`
	PreviewBytes int    `json:"preview_bytes,omitempty"`
}

// KwArg represents a keyword-value pair in a statement.
type KwArg struct {
	Keyword string `json:"kw"`
	Value   Value  `json:"value"`
}

// MarshalJSON implements custom JSON encoding for Value as required by product-guidelines.md.
func (v Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Kind Kind        `json:"kind"`
		V    interface{} `json:"v"`
	}{
		Kind: v.Kind,
		V:    v.V,
	})
}

// UnmarshalJSON implements custom JSON decoding for Value.
func (v *Value) UnmarshalJSON(data []byte) error {
	var raw struct {
		Kind Kind            `json:"kind"`
		V    json.RawMessage `json:"v"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	v.Kind = raw.Kind
	switch v.Kind {
	case KindInt:
		var i int
		if err := json.Unmarshal(raw.V, &i); err != nil {
			return err
		}
		v.V = i
	case KindBool:
		var b bool
		if err := json.Unmarshal(raw.V, &b); err != nil {
			return err
		}
		v.V = b
	case KindText:
		var th TextHandle
		if err := json.Unmarshal(raw.V, &th); err != nil {
			return err
		}
		v.V = th
	case KindJSON:
		var j interface{}
		if err := json.Unmarshal(raw.V, &j); err != nil {
			return err
		}
		v.V = j
	case KindString:
		var s string
		if err := json.Unmarshal(raw.V, &s); err != nil {
			return err
		}
		v.V = s
	case KindSpan:
		var s Span
		if err := json.Unmarshal(raw.V, &s); err != nil {
			return err
		}
		v.V = s
	case KindList:
		var l []Value
		if err := json.Unmarshal(raw.V, &l); err != nil {
			return err
		}
		v.V = l
	case KindNull:
		v.V = nil
	case KindOffset:
		var i int
		if err := json.Unmarshal(raw.V, &i); err != nil {
			return err
		}
		v.V = i
	default:
		return fmt.Errorf("unknown value kind: %s", v.Kind)
	}
	return nil
}
