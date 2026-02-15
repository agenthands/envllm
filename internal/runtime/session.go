package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/agenthands/envllm/internal/ast"
)

type BudgetExceededError struct {
	Message string
}

func (e *BudgetExceededError) Error() string { return e.Message }

type CapabilityDeniedError struct {
	Message string
}

func (e *CapabilityDeniedError) Error() string { return e.Message }

// TextStore interface.
type TextStore interface {
	Add(text string) TextHandle
	Get(h TextHandle) (string, bool)
	Window(h TextHandle, center, radius int) (TextHandle, error)
	Slice(h TextHandle, start, end int) (TextHandle, error)
}

// OpDispatcher allows the runtime to execute operations defined elsewhere.
type OpDispatcher interface {
	Dispatch(s *Session, name string, args []ast.KwArg) (Value, error)
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
	MaxStmtsPerCell     int             `json:"max_stmts_per_cell"`
	MaxWallTime         time.Duration   `json:"max_wall_time"`
	MaxTotalBytes       int             `json:"max_total_bytes"`
	MaxRecursionDepth   int             `json:"max_recursion_depth"`
	MaxSubcalls         int             `json:"max_subcalls"`
	AllowedCapabilities map[string]bool `json:"allowed_capabilities"`
	AllowedReadPaths    []string        `json:"allowed_read_paths"`
	AllowedWritePaths   []string        `json:"allowed_write_paths"`
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

	// Current execution context
	CurrentCell string
	CellIndex   int

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
		Cell: CellInfo{
			Name:  s.CurrentCell,
			Index: s.CellIndex,
		},
		Status:    status,
		VarsDelta: s.VarsDelta,
		Final:     s.Final,
		Events:    s.Events,
		Errors:    errors,
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
	s.CurrentCell = cell.Name
	
	for _, stmt := range cell.Stmts {
		if err := s.ExecuteStmt(ctx, stmt); err != nil {
			return err
		}
		
		// Check budgets
		if s.Policy.MaxStmtsPerCell > 0 && s.StmtsExecuted > s.Policy.MaxStmtsPerCell {
			return &BudgetExceededError{Message: fmt.Sprintf("max statements per cell (%d) exceeded", s.Policy.MaxStmtsPerCell)}
		}
		if s.Policy.MaxWallTime > 0 && time.Since(s.StartTime) > s.Policy.MaxWallTime {
			return &BudgetExceededError{Message: fmt.Sprintf("max wall time (%v) exceeded", s.Policy.MaxWallTime)}
		}
	}
	
	return nil
}

// ExecuteStmt runs a single statement.
func (s *Session) ExecuteStmt(ctx context.Context, stmt ast.Stmt) error {
	s.StmtsExecuted++

	switch st := stmt.(type) {
	case *ast.SetFinalStmt:
		val, err := s.EvalExpr(st.Source)
		if err != nil {
			return err
		}
		s.Final = &val
	case *ast.PrintStmt:
		val, err := s.EvalExpr(st.Source)
		if err != nil {
			return err
		}
		fmt.Printf("[PRINT] %v\n", val.V)
	case *ast.AssertStmt:
		val, err := s.EvalExpr(st.Cond)
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
		
		res, err := s.Dispatcher.Dispatch(s, st.OpName, st.Args)
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

func (s *Session) EvalExpr(expr ast.Expr) (Value, error) {
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
	case *ast.NullExpr:
		return Value{Kind: KindNull, V: nil}, nil
	default:
		return Value{}, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// ResolveIdent returns the name of the identifier if the expression is an IdentExpr.
func (s *Session) ResolveIdent(expr ast.Expr) (string, bool) {
	if e, ok := expr.(*ast.IdentExpr); ok {
		return e.Name, true
	}
	return "", false
}
