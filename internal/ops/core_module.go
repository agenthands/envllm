package ops

import (
	"context"
	"fmt"

	"github.com/agenthands/envllm/internal/ops/capability"
	"github.com/agenthands/envllm/internal/ops/pure"
	"github.com/agenthands/envllm/internal/runtime"
)

type CoreModule struct{}

func (m *CoreModule) ID() string { return "core" }

func (m *CoreModule) Operations() []Op {
	return []Op{
		{Name: "STATS", Capabilities: []string{"pure"}, ResultType: runtime.KindJSON, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindText}}, Into: true},
		{Name: "FIND_TEXT", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "NEEDLE", Type: runtime.KindText},
			{Kw: "MODE", Enum: []string{"FIRST", "LAST"}},
			{Kw: "IGNORE_CASE", Type: runtime.KindBool},
		}, Into: true},
		{Name: "WINDOW_TEXT", Capabilities: []string{"pure"}, ResultType: runtime.KindText, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "CENTER", Type: runtime.KindOffset},
			{Kw: "RADIUS", Type: runtime.KindInt},
		}, Into: true},
		{Name: "SLICE_TEXT", Capabilities: []string{"pure"}, ResultType: runtime.KindText, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "START", Type: runtime.KindOffset},
			{Kw: "END", Type: runtime.KindOffset},
		}, Into: true},
		{Name: "FIND_REGEX", Capabilities: []string{"pure"}, ResultType: runtime.KindSpan, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "PATTERN", Type: runtime.KindText},
			{Kw: "MODE", Enum: []string{"FIRST", "LAST"}},
		}, Into: true},
		{Name: "JSON_PARSE", Capabilities: []string{"pure"}, ResultType: runtime.KindJSON, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindText}}, Into: true},
		{Name: "JSON_GET", Capabilities: []string{"pure"}, ResultType: runtime.KindJSON, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindJSON},
			{Kw: "PATH", Type: runtime.KindText},
		}, Into: true},
		{Name: "GET_SPAN_START", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindSpan}}, Into: true},
		{Name: "GET_SPAN_END", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindSpan}}, Into: true},
		{Name: "CONCAT_TEXT", Capabilities: []string{"pure"}, ResultType: runtime.KindText, Signature: []Param{{Kw: "A", Type: runtime.KindText}, {Kw: "B", Type: runtime.KindText}}, Into: true},
		{Name: "TO_TEXT", Capabilities: []string{"pure"}, ResultType: runtime.KindText, Signature: []Param{{Kw: "VALUE", Type: ""}}, Into: true},
		{Name: "OFFSET", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{{Kw: "VALUE", Type: runtime.KindInt}}, Into: true},
		{Name: "SPAN", Capabilities: []string{"pure"}, ResultType: runtime.KindSpan, Signature: []Param{{Kw: "START", Type: runtime.KindOffset}, {Kw: "END", Type: runtime.KindOffset}}, Into: true},
		{Name: "SUBCALL", Capabilities: []string{"llm"}, ResultType: runtime.KindJSON, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "TASK", Type: runtime.KindText},
			{Kw: "DEPTH_COST", Type: runtime.KindInt},
		}, Into: true},
	}
}

func (m *CoreModule) Handlers() map[string]OpImplementation {
	return map[string]OpImplementation{
		"STATS": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.Stats(s, args[0])
		},
		"FIND_TEXT": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			mode := "FIRST"
			if m, ok := args[2].V.(string); ok { mode = m }
			ignoreCase := false
			if ic, ok := args[3].V.(bool); ok { ignoreCase = ic }
			return pure.FindText(s, args[0], args[1], mode, ignoreCase)
		},
		"WINDOW_TEXT": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.WindowText(s, args[0], args[1].V.(int), args[2].V.(int))
		},
		"SLICE_TEXT": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.SliceText(s, args[0], args[1].V.(int), args[2].V.(int))
		},
		"FIND_REGEX": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			mode := "FIRST"
			if m, ok := args[2].V.(string); ok { mode = m }
			return pure.FindRegex(s, args[0], args[1], mode)
		},
		"JSON_PARSE": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.JSONParse(s, args[0])
		},
		"JSON_GET": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			path := ""
			if p, ok := args[1].V.(string); ok { path = p } else if h, ok := args[1].V.(runtime.TextHandle); ok { path, _ = s.Stores.Text.Get(h) }
			return pure.JSONGet(s, args[0], path)
		},
		"GET_SPAN_START": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.GetSpanStart(s, args[0])
		},
		"GET_SPAN_END": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.GetSpanEnd(s, args[0])
		},
		"CONCAT_TEXT": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.ConcatText(s, args[0], args[1])
		},
		"TO_TEXT": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.ToText(s, args[0])
		},
		"OFFSET": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.Offset(s, args[0].V.(int))
		},
		"SPAN": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.Span(s, args[0].V.(int), args[1].V.(int))
		},
		"SUBCALL": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			if s.Host == nil { return runtime.Value{}, fmt.Errorf("SUBCALL failed: no host configured") }
			source := args[0].V.(runtime.TextHandle)
			var task string
			if args[1].Kind == runtime.KindString { task = args[1].V.(string) } else if args[1].Kind == runtime.KindText {
				taskHandle := args[1].V.(runtime.TextHandle)
				var ok bool
				task, ok = s.Stores.Text.Get(taskHandle)
				if !ok { return runtime.Value{}, fmt.Errorf("SUBCALL failed: task text not found") }
			} else { return runtime.Value{}, fmt.Errorf("SUBCALL failed: TASK must be TEXT or STRING, got %s", args[1].Kind) }
			depthCost := args[2].V.(int)
			if s.Policy.MaxSubcalls > 0 && s.SubcallCount >= s.Policy.MaxSubcalls { return runtime.Value{}, &runtime.BudgetExceededError{Message: "max subcalls reached"} }
			if s.Policy.MaxRecursionDepth > 0 && s.RecursionDepth+depthCost > s.Policy.MaxRecursionDepth { return runtime.Value{}, &runtime.BudgetExceededError{Message: fmt.Sprintf("recursion depth limit reached (cost %d)", depthCost)} }
			req := runtime.SubcallRequest{Source: source, Task: task, DepthCost: depthCost, Budgets: make(map[string]int)}
			res, err := s.Host.Subcall(context.Background(), req)
			if err != nil { return runtime.Value{}, fmt.Errorf("host subcall failed: %v", err) }
			s.SubcallCount++; s.RecursionDepth += depthCost
			return res.Result, nil
		},
	}
}

type FSModule struct{}

func (m *FSModule) ID() string { return "fs" }

func (m *FSModule) Operations() []Op {
	return []Op{
		{Name: "READ_FILE", Capabilities: []string{"fs_read"}, ResultType: runtime.KindText, Signature: []Param{{Kw: "PATH", Type: runtime.KindText}}, Into: true},
		{Name: "WRITE_FILE", Capabilities: []string{"fs_write"}, ResultType: runtime.KindBool, Signature: []Param{{Kw: "PATH", Type: runtime.KindText}, {Kw: "SOURCE", Type: runtime.KindText}}, Into: true},
		{Name: "LIST_DIR", Capabilities: []string{"fs_read"}, ResultType: runtime.KindJSON, Signature: []Param{{Kw: "PATH", Type: runtime.KindText}}, Into: true},
	}
}

func (m *FSModule) Handlers() map[string]OpImplementation {
	return map[string]OpImplementation{
		"READ_FILE": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) { return capability.ReadFile(s, args[0]) },
		"WRITE_FILE": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) { return capability.WriteFile(s, args[0], args[1]) },
		"LIST_DIR": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) { return capability.ListDir(s, args[0]) },
	}
}
