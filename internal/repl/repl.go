package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/agenthands/envllm/internal/lex"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/internal/parse"
	"github.com/agenthands/envllm/internal/runtime"
	"github.com/agenthands/envllm/internal/store"
)

const PROMPT = "rlm> "

// Start starts the REPL.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	// Setup session
	ts := store.NewTextStore()
	tbl, err := ops.LoadTable("assets/ops.json")
	if err != nil {
		tbl, _ = ops.LoadTable("../../assets/ops.json")
	}
	reg := ops.NewRegistry(tbl)
	session := runtime.NewSession(runtime.Policy{MaxStmtsPerCell: 100}, ts)
	session.Dispatcher = reg

	fmt.Fprintln(out, "EnvLM REPL 0.1")
	fmt.Fprintln(out, "Type 'exit' to quit.")

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		if line == "exit" {
			return
		}

		if line == "" {
			continue
		}

		// Parse as a cell
		// Add CELL wrapper if not present
		src := line
		if !strings.HasPrefix(line, "CELL") && !strings.HasPrefix(line, "RLMDSL") {
			src = "CELL repl:\n  " + line + "\n"
		}

		l := lex.NewLexer("repl.rlm", src)
		p := parse.NewParser(l)
		prog, err := p.Parse()
		if err != nil {
			fmt.Fprintf(out, "Parse error: %v\n", err)
			continue
		}

		for _, cell := range prog.Cells {
			err = session.ExecuteCell(context.Background(), cell)
			status := "ok"
			var errs []runtime.Error
			if err != nil {
				status = "error"
				errs = append(errs, runtime.Error{Message: err.Error()})
			}

			res := session.GenerateResult(status, errs)
			output, _ := res.ToJSON()
			fmt.Fprintln(out, string(output))

			// Clear delta for next turn
			session.VarsDelta = make(map[string]runtime.Value)
		}
	}
}
