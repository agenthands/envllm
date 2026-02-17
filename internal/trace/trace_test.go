package trace

import (
	"strings"
	"testing"
)

func TestTraceStep_Redaction(t *testing.T) {
	redactor := NewRedactor("STRICT")
	
	step := TraceStep{
		Op: "AUTH",
		Inputs: map[string]interface{}{
			"password": "secret_password_123",
			"user": "admin",
		},
		Outputs: map[string]interface{}{
			"token": "session_token_abc",
		},
	}
	
	step.Redact(redactor)
	
	inputs := step.Inputs.(map[string]interface{})
	if !strings.HasPrefix(inputs["password"].(string), "[REDACTED") {
		t.Errorf("password was not redacted: %v", inputs["password"])
	}
	if !strings.HasPrefix(inputs["user"].(string), "[REDACTED") {
		t.Errorf("user field was not redacted: %v", inputs["user"])
	}
	
	outputs := step.Outputs.(map[string]interface{})
	if !strings.HasPrefix(outputs["token"].(string), "[REDACTED") {
		t.Errorf("token was not redacted: %v", outputs["token"])
	}
}

func TestJSONLSink(t *testing.T) {
	tmpFile := "trace_test.jsonl"
	sink, err := NewJSONLSink(tmpFile)
	if err != nil {
		t.Fatalf("failed to create sink: %v", err)
	}
	defer sink.Close()
	
	step := TraceStep{
		Decision: DecisionAccept,
		Phase: PhaseExec,
		Op: "TEST_OP",
	}
	
	if err := sink.Emit(step); err != nil {
		t.Errorf("emit failed: %v", err)
	}
}
