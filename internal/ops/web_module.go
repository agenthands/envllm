package ops

import (
	"github.com/agenthands/envllm/internal/runtime"
)

type WebModule struct{}

func (m *WebModule) ID() string { return "web" }

func (m *WebModule) Operations() []Op {
	return []Op{
		{
			Name:         "NAVIGATE",
			Capabilities: []string{"web.navigate"},
			ResultType:   runtime.KindBool,
			Signature: []Param{
				{Kw: "URL", Type: runtime.KindText},
			},
			Into: true,
		},
		{
			Name:         "CLICK",
			Capabilities: []string{"web.dom.query"},
			ResultType:   runtime.KindBool,
			Signature: []Param{
				{Kw: "SELECTOR", Type: runtime.KindText},
			},
			Into: true,
		},
		{
			Name:         "TYPE",
			Capabilities: []string{"web.dom.query"},
			ResultType:   runtime.KindBool,
			Signature: []Param{
				{Kw: "SELECTOR", Type: runtime.KindText},
				{Kw: "TEXT", Type: runtime.KindText},
			},
			Into: true,
		},
	}
}

func (m *WebModule) Handlers() map[string]OpImplementation {
	// For now, these are just mocks to demonstrate the registry
	return map[string]OpImplementation{
		"NAVIGATE": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return runtime.Value{Kind: runtime.KindBool, V: true}, nil
		},
		"CLICK": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return runtime.Value{Kind: runtime.KindBool, V: true}, nil
		},
		"TYPE": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return runtime.Value{Kind: runtime.KindBool, V: true}, nil
		},
	}
}
