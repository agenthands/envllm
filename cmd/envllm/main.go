package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	dfmt "github.com/agenthands/envllm/internal/fmt"
	"github.com/agenthands/envllm/internal/lint"
	"github.com/agenthands/envllm/internal/migrate"
	"github.com/agenthands/envllm/internal/ops"
	"github.com/agenthands/envllm/internal/repl"
	"github.com/agenthands/envllm/internal/runtime"
	"github.com/agenthands/envllm/internal/trace"
	"github.com/agenthands/envllm/pkg/envllm"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "run":
		run()
	case "repl":
		replCmd()
	case "validate":
		validate()
	case "fmt":
		fmtCmd()
	case "migrate":
		migrateCmd()
	case "check":
		checkCmd()
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: envllm <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  run <file>      Execute an RLMDSL script")
	fmt.Println("  repl            Start an interactive REPL")
	fmt.Println("  validate <file> Validate script syntax and ops")
	fmt.Println("  fmt <file>      Format script to canonical form")
	fmt.Println("  migrate <file>  Migrate v0.1 script to v0.2 STRICT")
	fmt.Println("  check <file>    Check script for v0.2 canonical errors")
	fmt.Println("  help            Show this help text")
}

func run() {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	maxStmts := runCmd.Int("max-stmts", 100, "Maximum statements per cell")
	timeout := runCmd.Duration("timeout", 0, "Maximum wall time for execution")
	modeStr := runCmd.String("mode", "compat", "Parser mode (compat or strict)")
	tracePath := runCmd.String("trace", "", "Path to emit JSONL trace certificates")

	if len(os.Args) < 3 {
		fmt.Println("Usage: envllm run <file> [flags]")
		runCmd.PrintDefaults()
		os.Exit(1)
	}

	filename := os.Args[2]
	runCmd.Parse(os.Args[3:])

	var sink trace.Sink
	if *tracePath != "" {
		var err error
		sink, err = trace.NewJSONLSink(*tracePath)
		if err != nil {
			fmt.Printf("Trace sink error: %v\n", err)
			os.Exit(1)
		}
		defer sink.Close()
	}

	mode := envllm.ModeCompat
	if *modeStr == "strict" {
		mode = envllm.ModeStrict
	}

	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	prog, err := envllm.Compile(filename, string(src), mode)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		if sink != nil {
			sink.Emit(trace.TraceStep{
				Phase:    trace.PhaseParse,
				Decision: trace.DecisionReject,
				Error:    &trace.TraceError{Code: "PARSE_ERROR", Message: err.Error()},
			})
		}
		os.Exit(1)
	}

	opt := envllm.ExecOptions{
		Policy: runtime.Policy{
			MaxStmtsPerCell: *maxStmts,
			MaxWallTime:     *timeout,
		},
		TextStore: envllm.NewTextStore(),
		TraceSink: sink,
	}

	res, err := prog.Execute(context.Background(), opt)
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		os.Exit(1)
	}

	output, _ := res.ToJSON()
	fmt.Println(string(output))
}

func replCmd() {
	repl.Start(os.Stdin, os.Stdout)
}

func validate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: envllm validate <file>")
		os.Exit(1)
	}

	filename := os.Args[2]
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	_, err = envllm.Compile(filename, string(src), envllm.ModeCompat)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Validation successful")
}

func fmtCmd() {
	fmtFlagSet := flag.NewFlagSet("fmt", flag.ExitOnError)
	modeStr := fmtFlagSet.String("mode", "strict", "Format mode (compat or strict)")

	if len(os.Args) < 3 {
		fmt.Println("Usage: envllm fmt <file> [flags]")
		fmtFlagSet.PrintDefaults()
		os.Exit(1)
	}
	filename := os.Args[2]
	fmtFlagSet.Parse(os.Args[3:])

	mode := envllm.ModeStrict
	if *modeStr == "compat" {
		mode = envllm.ModeCompat
	}

	src, _ := os.ReadFile(filename)
	prog, err := envllm.Compile(filename, string(src), mode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(dfmt.Format(prog.AST))
}

func migrateCmd() {
	migrateFlagSet := flag.NewFlagSet("migrate", flag.ExitOnError)
	from := migrateFlagSet.String("from", "v0.1", "Source version")
	to := migrateFlagSet.String("to", "v0.2", "Target version")

	if len(os.Args) < 3 {
		fmt.Println("Usage: envllm migrate <file> [flags]")
		migrateFlagSet.PrintDefaults()
		os.Exit(1)
	}
	filename := os.Args[2]
	migrateFlagSet.Parse(os.Args[3:])

	fmt.Printf("Migrating %s from %s to %s...\n", filename, *from, *to)

	src, _ := os.ReadFile(filename)
	
	tbl, _ := ops.LoadTable("assets/ops.json")
	canonical, report, err := migrate.Migrate(string(src), tbl)
	if err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("--- Migration Report ---")
	for _, change := range report.Changes {
		fmt.Printf("- %s\n", change)
	}
	fmt.Println("\n--- Canonical STRICT Form ---")
	fmt.Print(canonical)
}

func checkCmd() {
	checkFlagSet := flag.NewFlagSet("check", flag.ExitOnError)
	modeStr := checkFlagSet.String("mode", "strict", "Check mode (compat or strict)")

	if len(os.Args) < 3 {
		fmt.Println("Usage: envllm check <file> [flags]")
		checkFlagSet.PrintDefaults()
		os.Exit(1)
	}

	filename := os.Args[2]
	checkFlagSet.Parse(os.Args[3:])

	mode := envllm.ModeStrict
	if *modeStr == "compat" {
		mode = envllm.ModeCompat
	}

	src, _ := os.ReadFile(filename)

	prog, err := envllm.Compile(filename, string(src), mode)
	if err != nil {
		fmt.Printf("Parse Error (%s mode): %v\n", *modeStr, err)
		os.Exit(1)
	}

	tbl, _ := ops.LoadTable("assets/ops.json")
	lnt := lint.NewLinter(tbl)
	if mode == envllm.ModeStrict {
		lnt.WithMode(lint.ModeStrict)
	}
	errs := lnt.Lint(prog.AST)
	
	if len(errs) > 0 {
		fmt.Printf("Found %d canonical errors:\n", len(errs))
		for _, e := range errs {
			fmt.Printf("[%s] %s (Loc: %v)\n", e.Code, e.Message, e.Loc)
			if e.Hint != "" {
				fmt.Printf("  Hint: %s\n", e.Hint)
			}
		}
		os.Exit(1)
	}
	
	fmt.Println("Canonical check passed (STRICT mode)")
}
