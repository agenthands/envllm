package envllm

import (
	"context"
	"testing"

	"github.com/agenthands/envllm/internal/runtime"
)

func TestPublicAPI(t *testing.T) {
	src := `RLMDSL 0.1
CELL test:
  PRINT SOURCE "hello"
`
	prog, err := Compile("test.rlm", src)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	opt := ExecOptions{
		Policy: runtime.Policy{MaxStmtsPerCell: 10},
	}
	res, err := prog.Execute(context.Background(), opt)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if res.Status != "ok" {
		t.Errorf("expected status ok, got %s", res.Status)
	}
}
