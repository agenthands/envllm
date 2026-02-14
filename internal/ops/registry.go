package ops

import (
	"fmt"

	"github.com/agenthands/rlm-go/internal/ops/pure"
	"github.com/agenthands/rlm-go/internal/runtime"
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
		mode, _ := args[2].V.(string)
		ignoreCase, _ := args[3].V.(bool)
		return pure.FindText(s, args[0], args[1], mode, ignoreCase)
	}
	r.impls["WINDOW_TEXT"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.WindowText(s, args[0], args[1].V.(int), args[2].V.(int))
	}
	// JSON Ops
	r.impls["JSON_PARSE"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.JSONParse(s, args[0])
	}
	r.impls["JSON_GET"] = func(s *runtime.Session, args []runtime.Value) (runtime.Value, error) {
		return pure.JSONGet(s, args[0], args[1].V.(string))
	}
}

// Dispatch implements runtime.OpDispatcher.
func (r *Registry) Dispatch(s *runtime.Session, name string, args []runtime.KwArg) (runtime.Value, error) {
	// 1. Validate signature
	var vargs []ValidatedKwArg
	opDef, ok := r.Table.Ops[name]
	if !ok {
		return runtime.Value{}, fmt.Errorf("unknown operation: %s", name)
	}

	for i, arg := range args {
		param := opDef.Signature[i]
		val := arg.Value
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

	// 2. Lookup implementation
	impl, ok := r.impls[name]
	if !ok {
		return runtime.Value{}, fmt.Errorf("operation %q has no implementation", name)
	}

	// 3. Prepare positional args for implementation
	var posArgs []runtime.Value
	for _, v := range vargs {
		posArgs = append(posArgs, v.Value)
	}

	// 4. Execute
	res, err := impl(s, posArgs)
	if err != nil {
		return runtime.Value{}, err
	}

	// 5. Final type check
	if op.ResultType != "" && res.Kind != op.ResultType {
		return runtime.Value{}, fmt.Errorf("%s: result type mismatch: expected %s, got %s", name, op.ResultType, res.Kind)
	}

	return res, nil
}
