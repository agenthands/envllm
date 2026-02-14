package runtime

import (
	"fmt"
)

// Env manages variable storage and scope for an RLM session.
type Env struct {
	vars map[string]Value
}

func NewEnv() *Env {
	return &Env{
		vars: make(map[string]Value),
	}
}

// Define sets a variable value, enforcing single-assignment.
func (e *Env) Define(name string, val Value) error {
	if _, ok := e.vars[name]; ok {
		return fmt.Errorf("variable %q already defined (single-assignment enforced)", name)
	}
	e.vars[name] = val
	return nil
}

// Get retrieves a variable value.
func (e *Env) Get(name string) (Value, bool) {
	val, ok := e.vars[name]
	return val, ok
}

// Vars returns a copy of all variables in the environment.
func (e *Env) Vars() map[string]Value {
	copy := make(map[string]Value)
	for k, v := range e.vars {
		copy[k] = v
	}
	return copy
}
