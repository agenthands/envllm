package ops

// Module defines an extension package for EnvLLM.
type Module interface {
	ID() string
	Operations() []Op
	Handlers() map[string]OpImplementation
}

// OpDef is an alias for Op to match the plan's naming if preferred, 
// but we already have Op struct in table.go.

func (r *Registry) RegisterModule(m Module) error {
	for _, op := range m.Operations() {
		o := op
		r.Table.Ops[op.Name] = &o
	}
	for name, impl := range m.Handlers() {
		r.impls[name] = impl
	}
	return nil
}
