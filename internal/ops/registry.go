package ops

import (
	"fmt"

	"github.com/agenthands/envllm/internal/ast"
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
	r.RegisterModule(&CoreModule{})
	r.RegisterModule(&FSModule{})
	r.RegisterModule(&WebModule{})
	return r
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

	// fmt.Printf("[DEBUG] Dispatch %s, Caps: %v, Policy: %v\n", name, op.Capabilities, s.Policy.AllowedCapabilities)

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
