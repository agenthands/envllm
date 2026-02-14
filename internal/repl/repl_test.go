package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestREPL(t *testing.T) {
	input := `PRINT SOURCE "hello"
exit
`
	in := strings.NewReader(input)
	var out bytes.Buffer

	Start(in, &out)

	output := out.String()
	if !strings.Contains(output, "EnvLM REPL") {
		t.Errorf("expected header, got %q", output)
	}
	if !strings.Contains(output, `"status":"ok"`) {
		t.Errorf("expected ok status, got %q", output)
	}
}

func TestREPL_Persistence(t *testing.T) {
	input := `SET_FINAL SOURCE 42
exit
`
	in := strings.NewReader(input)
	var out bytes.Buffer

	Start(in, &out)

	output := out.String()
	if !strings.Contains(output, `"v":42`) {
		t.Errorf("expected final value 42, got %q", output)
	}
}

func TestREPL_EmptyLine(t *testing.T) {
	input := "\nexit\n"
	in := strings.NewReader(input)
	var out bytes.Buffer
	Start(in, &out)
	// Just ensuring it doesn't crash
}

func TestREPL_Error(t *testing.T) {
	input := "INVALID_STMT\nexit\n"
	in := strings.NewReader(input)
	var out bytes.Buffer
	Start(in, &out)
	output := out.String()
	if !strings.Contains(output, "error") {
		t.Errorf("expected error in output, got %q", output)
	}
}
