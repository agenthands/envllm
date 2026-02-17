package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/trace"
)

type BudgetExceededError struct {
	Message string
}

func (e *BudgetExceededError) Error() string { return e.Message }

type CapabilityDeniedError struct {
	Message string
}

func (e *CapabilityDeniedError) Error() string { return e.Message }

type ExtensionVersionUnsupportedError struct {
	Extension string
	Requested string
	Available string
}

func (e *ExtensionVersionUnsupportedError) Error() string {
	return fmt.Sprintf("ERR_EXTENSION_VERSION_UNSUPPORTED: %s (requested %s, available %s)", e.Extension, e.Requested, e.Available)
}

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
	// Trace Sink
	TraceSink trace.Sink
}

func (s *Session) emitTrace(step trace.TraceStep) {
	if s.TraceSink != nil {
		if step.Timestamp.IsZero() {
			step.Timestamp = time.Now()
		}
		if step.Phase == "" {
			step.Phase = trace.PhaseExec
		}
		s.TraceSink.Emit(step)
	}
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

// ExecuteTask runs a full task including inputs and body.
func (s *Session) ExecuteTask(ctx context.Context, task *ast.Task) error {
	s.StartTime = time.Now()
	// Inputs are assumed to be pre-set in Env for now, 
	// but we could validate them against task.Inputs here.
	
	if err := s.ExecuteBody(ctx, task.Body); err != nil {
		return err
	}
	
	// Final output check
	if task.Output != "" {
		val, ok := s.Env.Get(task.Output)
		if !ok {
			return fmt.Errorf("task output %q not found in environment", task.Output)
		}
		s.Final = &val
	}
	
	return nil
}

// ExecuteBody runs a sequence of body items.
func (s *Session) ExecuteBody(ctx context.Context, body []ast.BodyItem) error {
	for _, item := range body {
		if err := s.ExecuteBodyItem(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteBodyItem runs a single requirement, cell, or if statement.
func (s *Session) ExecuteBodyItem(ctx context.Context, item ast.BodyItem) error {
	switch it := item.(type) {
	case *ast.Requirement:
		// Requirements are currently just metadata for the linter/host.
		// Runtime can ignore them or validate against policy.
		return nil
	case *ast.Cell:
		return s.ExecuteCell(ctx, it)
	case *ast.IfStmt:
		return s.ExecuteIf(ctx, it)
	case ast.Stmt:
		return s.ExecuteStmt(ctx, it)
	default:
		return fmt.Errorf("unknown body item type: %T", item)
	}
}

// ExecuteIf runs an IF/ELSE block.
func (s *Session) ExecuteIf(ctx context.Context, stmt *ast.IfStmt) error {
	val, err := s.EvalExpr(stmt.Cond)
	if err != nil {
		return err
	}
	if val.Kind != KindBool {
		return fmt.Errorf("IF condition must be BOOL, got %s", val.Kind)
	}
	
	if val.V.(bool) {
		return s.ExecuteBody(ctx, stmt.ThenBody)
	} else if stmt.ElseBody != nil {
		return s.ExecuteBody(ctx, stmt.ElseBody)
	}
	return nil
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
	case *ast.ForEachStmt:
		collVal, ok := s.Env.Get(st.Collection)
		if !ok {
			return fmt.Errorf("undefined collection: %s", st.Collection)
		}
		if collVal.Kind != KindRows {
			return fmt.Errorf("FOR_EACH expects ROWS, got %s", collVal.Kind)
		}
		
		rows := collVal.V.([]map[string]interface{})
		limit := st.Limit
		if limit > len(rows) { limit = len(rows) }
		
		for i := 0; i < limit; i++ {
			// Define iterator variable (shadowing allowed inside loop for iterator only?)
			// Spec says: Loop variable is read-only.
			// We need a scoped environment or just define/undefine.
			// EnvLLM v0.1 has flat scope. v0.2.2 spec says "No nested FOR_EACH".
			// Let's implement scope push/pop for loop.
			
			// Actually, just define it. If it exists, error (NO_REUSE).
			// But it needs to exist for the loop.
			// Let's defer strict scope to Linter. Runtime just overwrites or defines.
			
			// Create a child scope/env? No, Env is flat map.
			// Hack: define, execute, undefine? Or allow reassignment for iterator?
			// The Linter will block reassignment. 
			// We should probably allow the iterator to be 'rebound' each iteration in the runtime.
			
			rowVal := Value{Kind: KindStruct, V: rows[i]}
			s.Env.vars[st.Iterator] = rowVal // Direct set to bypass single-assign check for loop
			
			for _, bs := range st.Body {
				if err := s.ExecuteStmt(ctx, bs); err != nil {
					return err
				}
			}
		}
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
			err := fmt.Errorf("no operation dispatcher configured")
			s.emitTrace(trace.TraceStep{
				Op:       st.OpName,
				Decision: trace.DecisionReject,
				Error:    &trace.TraceError{Code: "DISPATCH_ERROR", Message: err.Error()},
			})
			return err
		}
		
		res, err := s.Dispatcher.Dispatch(s, st.OpName, st.Args)
		if err != nil {
			s.emitTrace(trace.TraceStep{
				Op:       st.OpName,
				Decision: trace.DecisionReject,
				Error:    &trace.TraceError{Code: "EXEC_ERROR", Message: err.Error()},
			})
			return err
		}
		
		// Handle INTO
		if st.Into != "" {
			if err := s.defineVar(st.Into, res); err != nil {
				return err
			}
		}
		
		s.emitTrace(trace.TraceStep{
			Op:       st.OpName,
			Decision: trace.DecisionAccept,
			Outputs:  map[string]interface{}{st.Into: res.Kind},
		})

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
