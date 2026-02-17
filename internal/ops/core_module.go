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
		{Name: "STATS", Capabilities: []string{"pure"}, ResultType: runtime.KindStruct, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindText}}, Into: true},
		{Name: "GET_FIELD", Capabilities: []string{"pure"}, ResultType: runtime.KindJSON, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindStruct}, {Kw: "FIELD", Type: runtime.KindText}}, Into: true},
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
		{Name: "FIND_REGEX", Capabilities: []string{"pure"}, ResultType: runtime.KindStruct, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "PATTERN", Type: runtime.KindText},
			{Kw: "MODE", Enum: []string{"FIRST", "LAST"}},
		}, Into: true},
		{Name: "AFTER_TEXT", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "NEEDLE", Type: runtime.KindText},
			{Kw: "MODE", Enum: []string{"FIRST", "LAST"}},
			{Kw: "IGNORE_CASE", Type: runtime.KindBool},
		}, Into: true},
		{Name: "AFTER_REGEX", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "PATTERN", Type: runtime.KindText},
			{Kw: "MODE", Enum: []string{"FIRST", "LAST"}},
		}, Into: true},
		{Name: "MATCH_GROUP", Capabilities: []string{"pure"}, ResultType: runtime.KindSpan, Signature: []Param{
			{Kw: "MATCH", Type: runtime.KindStruct},
			{Kw: "INDEX", Type: runtime.KindInt},
		}, Into: true},
		{Name: "CAPTURE_REGEX_GROUP", Capabilities: []string{"pure"}, ResultType: runtime.KindSpan, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "PATTERN", Type: runtime.KindText},
			{Kw: "INDEX", Type: runtime.KindInt},
		}, Into: true},
		{Name: "VALUE_AFTER_DELIM", Capabilities: []string{"pure"}, ResultType: runtime.KindSpan, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "DELIM", Type: runtime.KindText},
			{Kw: "UNTIL", Type: runtime.KindText},
		}, Into: true},
		{Name: "EXTRACT_JSON", Capabilities: []string{"pure"}, ResultType: runtime.KindJSON, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
		}, Into: true},
		{Name: "EXTRACT_VALUE", Capabilities: []string{"pure"}, ResultType: runtime.KindText, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindText},
			{Kw: "KEY", Type: runtime.KindText},
			{Kw: "UNTIL", Type: runtime.KindText},
		}, Into: true},
		{Name: "JSON_PARSE", Capabilities: []string{"pure"}, ResultType: runtime.KindJSON, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindText}}, Into: true},
		{Name: "JSON_GET", Capabilities: []string{"pure"}, ResultType: runtime.KindJSON, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindJSON},
			{Kw: "PATH", Type: runtime.KindText},
		}, Into: true},
		{Name: "SELECT_FIELDS", Capabilities: []string{"pure"}, ResultType: runtime.KindRows, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindRows},
			{Kw: "FIELDS", Type: runtime.KindList},
		}, Into: true},
		{Name: "FILTER_ROWS", Capabilities: []string{"pure"}, ResultType: runtime.KindRows, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindRows},
			{Kw: "KEY", Type: runtime.KindText},
			{Kw: "OP", Enum: []string{"==", "!=", ">", "<"}},
			{Kw: "VALUE", Type: ""},
		}, Into: true},
		{Name: "AGGREGATE_ROWS", Capabilities: []string{"pure"}, ResultType: runtime.KindRows, Signature: []Param{
			{Kw: "SOURCE", Type: runtime.KindRows},
			{Kw: "GROUP_BY", Type: runtime.KindText},
			{Kw: "COMPUTE", Enum: []string{"COUNT", "SUM", "AVG"}},
		}, Into: true},
		{Name: "GET_SPAN_START", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindSpan}}, Into: true},
		{Name: "GET_SPAN_END", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{{Kw: "SOURCE", Type: runtime.KindSpan}}, Into: true},
		{Name: "CONCAT_TEXT", Capabilities: []string{"pure"}, ResultType: runtime.KindText, Signature: []Param{{Kw: "A", Type: runtime.KindText}, {Kw: "B", Type: runtime.KindText}}, Into: true},
		{Name: "TO_TEXT", Capabilities: []string{"pure"}, ResultType: runtime.KindText, Signature: []Param{{Kw: "VALUE", Type: ""}}, Into: true},
		{Name: "OFFSET", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{{Kw: "VALUE", Type: runtime.KindInt}}, Into: true},
		{Name: "OFFSET_ADD", Capabilities: []string{"pure"}, ResultType: runtime.KindOffset, Signature: []Param{{Kw: "OFFSET", Type: runtime.KindOffset}, {Kw: "AMOUNT", Type: runtime.KindInt}}, Into: true},
		{Name: "SPAN", Capabilities: []string{"pure"}, ResultType: runtime.KindSpan, Signature: []Param{{Kw: "START", Type: runtime.KindOffset}, {Kw: "END", Type: runtime.KindOffset}}, Into: true},
		{Name: "AS_SPAN", Capabilities: []string{"pure"}, ResultType: runtime.KindSpan, Signature: []Param{{Kw: "OFFSET", Type: runtime.KindOffset}, {Kw: "LEN", Type: runtime.KindInt}}, Into: true},
		{Name: "GET_COST", Capabilities: []string{"pure"}, ResultType: runtime.KindCost, Signature: []Param{{Kw: "RESULT", Type: runtime.KindJSON}}, Into: true},
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
		"GET_FIELD": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			// Extract field name from TEXT handle
			h := args[1].V.(runtime.TextHandle)
			field, _ := s.Stores.Text.Get(h)
			return pure.GetField(s, args[0], field)
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
		"AFTER_TEXT": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			mode := "FIRST"
			if m, ok := args[2].V.(string); ok { mode = m }
			ignoreCase := false
			if ic, ok := args[3].V.(bool); ok { ignoreCase = ic }
			return pure.AfterText(s, args[0], args[1], mode, ignoreCase)
		},
		"AFTER_REGEX": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			mode := "FIRST"
			if m, ok := args[2].V.(string); ok { mode = m }
			return pure.AfterRegex(s, args[0], args[1], mode)
		},
		"MATCH_GROUP": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.MatchGroup(s, args[0], args[1].V.(int))
		},
		"CAPTURE_REGEX_GROUP": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.CaptureRegexGroup(s, args[0], args[1], args[2].V.(int))
		},
		"VALUE_AFTER_DELIM": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.ValueAfterDelim(s, args[0], args[1], args[2])
		},
		"EXTRACT_JSON": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.ExtractJSON(s, args[0])
		},
		"EXTRACT_VALUE": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.ExtractValue(s, args[0], args[1], args[2])
		},
		"JSON_PARSE": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.JSONParse(s, args[0])
		},
		"JSON_GET": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			path := ""
			if p, ok := args[1].V.(string); ok { path = p } else if h, ok := args[1].V.(runtime.TextHandle); ok { path, _ = s.Stores.Text.Get(h) }
			return pure.JSONGet(s, args[0], path)
		},
		"SELECT_FIELDS": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.SelectFields(s, args[0], args[1])
		},
		"FILTER_ROWS": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			h := args[1].V.(runtime.TextHandle)
			key, _ := s.Stores.Text.Get(h)
			op := args[2].V.(string)
			return pure.FilterRows(s, args[0], key, op, args[3])
		},
		"AGGREGATE_ROWS": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			h := args[1].V.(runtime.TextHandle)
			groupBy, _ := s.Stores.Text.Get(h)
			compute := args[2].V.(string)
			return pure.AggregateRows(s, args[0], groupBy, compute)
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
		"OFFSET_ADD": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.OffsetAdd(s, args[0], args[1].V.(int))
		},
		"SPAN": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.Span(s, args[0].V.(int), args[1].V.(int))
		},
		"AS_SPAN": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.AsSpan(s, args[0].V.(int), args[1].V.(int))
		},
		"GET_COST": func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
			return pure.GetCost(s, args[0])
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
