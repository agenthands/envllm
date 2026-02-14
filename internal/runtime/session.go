package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/agenthands/rlm-go/internal/ast"
)

// TextStore interface.
type TextStore interface {
	Add(text string) TextHandle
	Get(h TextHandle) (string, bool)
	Window(h TextHandle, center, radius int) (TextHandle, error)
}

// OpDispatcher allows the runtime to execute operations defined elsewhere.
type OpDispatcher interface {
	Dispatch(s *Session, name string, args []KwArg) (Value, error)
}

// Host interface defines the interaction between the runtime and the LLM environment.
type Host interface {
	Subcall(ctx context.Context, req SubcallRequest) (SubcallResponse, error)
}

// SubcallRequest represents a request sent to the Host.
type SubcallRequest struct {
	Source    TextHandle
	Task      string
	DepthCost int
	Budgets   map[string]int // Inherited/Passthrough budgets
}

// SubcallResponse represents the result returned by the Host.
type SubcallResponse struct {
	Result Value
	Stats  map[string]int
}

// Policy defines the resource limits for an RLM session.
type Policy struct {
	MaxStmtsPerCell     int
	MaxWallTime         time.Duration
	MaxTotalBytes       int
	MaxRecursionDepth   int
	MaxSubcalls         int
	AllowedCapabilities map[string]bool
	AllowedReadPaths    []string
	AllowedWritePaths   []string
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
	StmtsExecuted  int
	StartTime      time.Time
	RecursionDepth int
	SubcallCount   int

	// Result tracking
	Events    []Event
	VarsDelta map[string]Value

	// Dispatcher
	Dispatcher OpDispatcher
	// Host
	Host Host
}

func NewSession(policy Policy, ts TextStore) *Session {
	s := &Session{
		Env:            NewEnv(),
		Policy:         policy,
		VarsDelta:      make(map[string]Value),
		RecursionDepth: 0,
		SubcallCount:   0,
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

// ValidatePath ensures a path is within the whitelist for read or write mode.
func (s *Session) ValidatePath(path string, write bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %v", err)
	}

	whitelist := s.Policy.AllowedReadPaths
	if write {
		whitelist = s.Policy.AllowedWritePaths
	}

	for _, wp := range whitelist {
		absWhitelist, err := filepath.Abs(wp)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absWhitelist) {
			return nil
		}
	}

	mode := "read"
	if write {
		mode = "write"
	}
	return fmt.Errorf("security_error: %s access to %q denied by policy", mode, path)
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
	res.Budgets["recursion_depth"] = BudgetStats{Used: s.RecursionDepth, Limit: s.Policy.MaxRecursionDepth}
	res.Budgets["subcalls"] = BudgetStats{Used: s.SubcallCount, Limit: s.Policy.MaxSubcalls}
	
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
		if s.Dispatcher == nil {
			return fmt.Errorf("no operation dispatcher configured")
		}
		
		// Evaluate arguments
		var args []KwArg
		for _, arg := range st.Args {
			val, err := s.evalExpr(arg.Value)
			if err != nil {
				return err
			}
			args = append(args, KwArg{Keyword: arg.Keyword, Value: val})
		}
		
		res, err := s.Dispatcher.Dispatch(s, st.OpName, args)
		if err != nil {
			return err
		}
		
		// Handle INTO
		if st.Into != "" {
			if err := s.defineVar(st.Into, res); err != nil {
				return err
			}
		}
		
		// Record event
		s.Events = append(s.Events, Event{
			T:    "op",
			Op:   st.OpName,
			Into: st.Into,
		})

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
		return Value{Kind: KindString, V: e.Value}, nil
	case *ast.IntExpr:
		return Value{Kind: KindInt, V: e.Value}, nil
	case *ast.BoolExpr:
		return Value{Kind: KindBool, V: e.Value}, nil
	default:
		return Value{}, fmt.Errorf("unknown expression type: %T", expr)
	}
}
