package ops

import (
	"context"
	"fmt"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/ops/capability"
	"github.com/agenthands/envllm/internal/ops/pure"
	"github.com/agenthands/envllm/internal/runtime"
)

// OpImplementation is the function signature for operation logic.
type OpImplementation func(s *runtime.Session, args []runtime.Value) (runtime.Value, error)

// Registry maps operation names to their implementations and metadata.
type Registry struct {
	Table *Table
	impls map[string]OpImplementation
}

func NewRegistry(tbl *Table) *Registry {
	r := &Registry{
		Table: tbl,
		impls: make(map[string]OpImplementation),
	}
	r.registerDefaults()
	return r
}

func (r *Registry) registerDefaults() {
	// Text Ops
	r.impls["STATS"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.Stats(s, args[0])
	}
	r.impls["FIND_TEXT"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		// Table validation ensures these V types are correct based on Kind
		mode, _ := args[2].V.(string)
		ignoreCase, _ := args[3].V.(bool)
		return pure.FindText(s, args[0], args[1], mode, ignoreCase)
	}
	r.impls["WINDOW_TEXT"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.WindowText(s, args[0], args[1].V.(int), args[2].V.(int))
	}
	r.impls["SLICE_TEXT"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.SliceText(s, args[0], args[1].V.(int), args[2].V.(int))
	}
	r.impls["FIND_REGEX"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.FindRegex(s, args[0], args[1], args[2].V.(string))
	}
	// JSON Ops
	r.impls["JSON_PARSE"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.JSONParse(s, args[0])
	}
	r.impls["JSON_GET"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.JSONGet(s, args[0], args[1].V.(string))
	}
	// Recursive Ops
	r.impls["SUBCALL"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		if s.Host == nil {
			return runtime.Value{}, fmt.Errorf("SUBCALL failed: no host configured")
		}

		source := args[0].V.(runtime.TextHandle)
		taskHandle := args[1].V.(runtime.TextHandle)
		depthCost := args[2].V.(int)

		// Get task string from handle
		task, ok := s.Stores.Text.Get(taskHandle)
		if !ok {
			return runtime.Value{}, fmt.Errorf("SUBCALL failed: task text not found")
		}

		// Validate budgets
		if s.Policy.MaxSubcalls > 0 && s.SubcallCount >= s.Policy.MaxSubcalls {
			return runtime.Value{}, &runtime.BudgetExceededError{Message: "max subcalls reached"}
		}
		if s.Policy.MaxRecursionDepth > 0 && s.RecursionDepth+depthCost > s.Policy.MaxRecursionDepth {
			return runtime.Value{}, &runtime.BudgetExceededError{Message: fmt.Sprintf("recursion depth limit reached (cost %d)", depthCost)}
		}

		// Prepare request
		req := runtime.SubcallRequest{
			Source:    source,
			Task:      task,
			DepthCost: depthCost,
			Budgets:   make(map[string]int), // TODO: populate with remaining budgets
		}

		// Execute call
		res, err := s.Host.Subcall(context.Background(), req)
		if err != nil {
			return runtime.Value{}, fmt.Errorf("host subcall failed: %v", err)
		}

		// Update stats
		s.SubcallCount++
		s.RecursionDepth += depthCost

		return res.Result, nil
	}
	// FS Ops
	r.impls["READ_FILE"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return capability.ReadFile(s, args[0])
	}
	r.impls["WRITE_FILE"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return capability.WriteFile(s, args[0], args[1])
	}
	r.impls["LIST_DIR"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return capability.ListDir(s, args[0])
	}
	r.impls["GET_SPAN_START"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.GetSpanStart(s, args[0])
	}
	r.impls["GET_SPAN_END"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.GetSpanEnd(s, args[0])
	}
}

// Dispatch implements runtime.OpDispatcher.
func (r *Registry) Dispatch(s *runtime.Session, name string, args []ast.KwArg) (runtime.Value, error) {
	// 1. Validate signature and evaluate args
	var vargs []ValidatedKwArg
	opDef, ok := r.Table.Ops[name]
	if !ok {
		return runtime.Value{}, fmt.Errorf("unknown operation: %s", name)
	}

	if len(args) != len(opDef.Signature) {
		return runtime.Value{}, fmt.Errorf("%s: expected %d arguments, got %d", name, len(opDef.Signature), len(args))
	}

	for i, arg := range args {
		param := opDef.Signature[i]
		
		var val runtime.Value
		var err error

		// If it's an enum, we allow raw identifiers as strings
		if len(param.Enum) > 0 {
			if name, ok := s.ResolveIdent(arg.Value); ok {
				val = runtime.Value{Kind: runtime.KindString, V: name}
			}
		}

		// If not already resolved as enum identifier, evaluate it
		if val.Kind == "" {
			val, err = s.EvalExpr(arg.Value)
			if err != nil {
				return runtime.Value{}, err
			}
		}

		// Promote STRING to TEXT if needed
		if param.Type == runtime.KindText && val.Kind == runtime.KindString && s.Stores.Text != nil {
			h := s.Stores.Text.Add(val.V.(string))
			val = runtime.Value{Kind: runtime.KindText, V: h}
		}
		vargs = append(vargs, ValidatedKwArg{Keyword: arg.Keyword, Value: val})
	}
	
	op, err := r.Table.ValidateSignature(name, vargs)
	if err != nil {
		return runtime.Value{}, err
	}

	// 2. Check capabilities
	for _, cap := range op.Capabilities {
		if cap == "pure" {
			continue
		}
		if s.Policy.AllowedCapabilities == nil || !s.Policy.AllowedCapabilities[cap] {
			return runtime.Value{}, &runtime.CapabilityDeniedError{Message: fmt.Sprintf("capability %q denied by policy", cap)}
		}
	}

	// 3. Lookup implementation
	impl, ok := r.impls[name]
	if !ok {
		return runtime.Value{}, fmt.Errorf("operation %q has no implementation", name)
	}

	// 4. Prepare positional args for implementation
	var posArgs []runtime.Value
	for _, v := range vargs {
		posArgs = append(posArgs, v.Value)
	}

	// 5. Execute
	res, err := impl(s, posArgs)
	if err != nil {
		return runtime.Value{}, err
	}

	// 6. Final type check
	if op.ResultType != "" && res.Kind != op.ResultType {
		return runtime.Value{}, fmt.Errorf("%s: result type mismatch: expected %s, got %s", name, op.ResultType, res.Kind)
	}

	return res, nil
}
