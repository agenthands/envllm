package trace

import (
	"time"
)

type Decision string

const (
	DecisionAccept Decision = "accept"
	DecisionReject Decision = "reject"
)

type Phase string

const (
	PhaseParse      Phase = "parse"
	PhaseLint       Phase = "lint"
	PhaseTypeCheck  Phase = "typecheck"
	PhaseCapability Phase = "capability"
	PhaseExec       Phase = "exec"
)

// TraceStep represents a single recorded step in the program lifecycle.
type TraceStep struct {
	TraceVersion string      `json:"trace_version"`
	RunID        string      `json:"run_id"`
	ProgramID    string      `json:"program_id"`
	Phase        Phase       `json:"phase"`
	Seq          int         `json:"seq"`
	StepID       string      `json:"step_id,omitempty"`
	Op           string      `json:"op,omitempty"`
	Inputs       interface{} `json:"inputs,omitempty"`
	Expected     interface{} `json:"expected,omitempty"`
	Actual       interface{} `json:"actual,omitempty"`
	Outputs      interface{} `json:"outputs,omitempty"`
	Decision     Decision    `json:"decision"`
	Error        *TraceError `json:"error,omitempty"`
	RuleID       string      `json:"rule_id,omitempty"`
	Hint         string      `json:"hint,omitempty"`
	Timestamp    time.Time   `json:"ts"`
}

type TraceError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Redact applies masking to the TraceStep fields.
func (s *TraceStep) Redact(r *Redactor) {
	if r == nil {
		return
	}
	s.Inputs = r.Redact(s.Inputs)
	s.Outputs = r.Redact(s.Outputs)
	if s.Error != nil {
		s.Error.Details = r.Redact(s.Error.Details)
	}
}
