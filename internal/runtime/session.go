package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/agenthands/rlm-go/internal/ast"
)

// TextStore interface to avoid circular dependency if needed, 
// but since store is in internal/store, and it imports runtime, 
// we have a circular dependency if runtime imports store.
// Let's define the interface here.
type TextStore interface {
	Add(text string) TextHandle
	Get(h TextHandle) (string, bool)
}

// Policy defines the resource limits for an RLM session.
type Policy struct {
	MaxStmtsPerCell int
	MaxWallTime     time.Duration
	MaxTotalBytes   int
}

// Session represents an active RLM session.
type Session struct {
	Env    *Env
	Stores struct {
		Text TextStore
	}
	Policy Policy
	Final  *Value
	
	// Stats for budgeting
	StmtsExecuted int
	StartTime     time.Time

	// Result tracking
	Events    []Event
	VarsDelta map[string]Value
}

func NewSession(policy Policy, ts TextStore) *Session {
	s := &Session{
		Env:       NewEnv(),
		Policy:    policy,
		VarsDelta: make(map[string]Value),
	}
	s.Stores.Text = ts
	return s
}

func (s *Session) defineVar(name string, val Value) error {
	if err := s.Env.Define(name, val); err != nil {
		return err
	}
	s.VarsDelta[name] = val
	return nil
}

func (s *Session) GenerateResult(status string, errors []Error) ExecResult {
	res := ExecResult{
		SchemaVersion: "obs-0.1",
		Status:        status,
		VarsDelta:     s.VarsDelta,
		Final:         s.Final,
		Events:        s.Events,
		Errors:        errors,
	}
	
	// Add budgets
	res.Budgets = make(map[string]BudgetStats)
	res.Budgets["stmts"] = BudgetStats{Used: s.StmtsExecuted, Limit: s.Policy.MaxStmtsPerCell}
	if s.Policy.MaxWallTime > 0 {
		res.Budgets["wall_time_ms"] = BudgetStats{
			Used:  int(time.Since(s.StartTime).Milliseconds()),
			Limit: int(s.Policy.MaxWallTime.Milliseconds()),
		}
	}
	
	return res
}

// ExecuteCell runs all statements in a cell.
func (s *Session) ExecuteCell(ctx context.Context, cell *ast.Cell) error {
	s.StartTime = time.Now()
	
	for _, stmt := range cell.Stmts {
		if err := s.ExecuteStmt(ctx, stmt); err != nil {
			return err
		}
		
		// Check budgets
		if s.Policy.MaxStmtsPerCell > 0 && s.StmtsExecuted > s.Policy.MaxStmtsPerCell {
			return fmt.Errorf("budget exceeded: max statements per cell (%d)", s.Policy.MaxStmtsPerCell)
		}
		if s.Policy.MaxWallTime > 0 && time.Since(s.StartTime) > s.Policy.MaxWallTime {
			return fmt.Errorf("budget exceeded: max wall time (%v)", s.Policy.MaxWallTime)
		}
	}
	
	return nil
}

// ExecuteStmt runs a single statement.
func (s *Session) ExecuteStmt(ctx context.Context, stmt ast.Stmt) error {
	s.StmtsExecuted++

	switch st := stmt.(type) {
	case *ast.SetFinalStmt:
		val, err := s.evalExpr(st.Source)
		if err != nil {
			return err
		}
		s.Final = &val
	case *ast.PrintStmt:
		val, err := s.evalExpr(st.Source)
		if err != nil {
			return err
		}
		// In v0.1 we just trace it or print to stdout for now.
		fmt.Printf("[PRINT] %v\n", val.V)
	case *ast.AssertStmt:
		val, err := s.evalExpr(st.Cond)
		if err != nil {
			return err
		}
		if val.Kind != KindBool {
			return fmt.Errorf("%s: ASSERT COND must be BOOL, got %s", st.Pos(), val.Kind)
		}
		if !val.V.(bool) {
			return fmt.Errorf("assertion failed: %s", st.Message)
		}
	case *ast.OpStmt:
		// Basic OpStmt handling (to be expanded in next track)
		return fmt.Errorf("operation %q not yet implemented", st.OpName)
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
	
	return nil
}

func (s *Session) evalExpr(expr ast.Expr) (Value, error) {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		val, ok := s.Env.Get(e.Name)
		if !ok {
			return Value{}, fmt.Errorf("undefined variable: %s", e.Name)
		}
		return val, nil
	case *ast.StringExpr:
		if s.Stores.Text == nil {
			return Value{Kind: KindJSON, V: e.Value}, nil
		}
		h := s.Stores.Text.Add(e.Value)
		return Value{Kind: KindText, V: h}, nil
	case *ast.IntExpr:
		return Value{Kind: KindInt, V: e.Value}, nil
	case *ast.BoolExpr:
		return Value{Kind: KindBool, V: e.Value}, nil
	default:
		return Value{}, fmt.Errorf("unknown expression type: %T", expr)
	}
}
