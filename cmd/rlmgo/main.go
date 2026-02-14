package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/agenthands/rlm-go/internal/repl"
	"github.com/agenthands/rlm-go/internal/runtime"
	"github.com/agenthands/rlm-go/pkg/rlmgo"
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
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: rlmgo <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  run <file>      Execute an RLMDSL script")
	fmt.Println("  repl            Start an interactive REPL")
	fmt.Println("  validate <file> Validate script syntax and ops")
	fmt.Println("  help            Show this help text")
}

func run() {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	maxStmts := runCmd.Int("max-stmts", 100, "Maximum statements per cell")
	timeout := runCmd.Duration("timeout", 0, "Maximum wall time for execution")
	
	if len(os.Args) < 3 {
		fmt.Println("Usage: rlmgo run <file> [flags]")
		runCmd.PrintDefaults()
		os.Exit(1)
	}

	filename := os.Args[2]
	runCmd.Parse(os.Args[3:])

	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	prog, err := rlmgo.Compile(filename, string(src))
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		os.Exit(1)
	}

	opt := rlmgo.ExecOptions{
		Policy: runtime.Policy{
			MaxStmtsPerCell: *maxStmts,
			MaxWallTime:     *timeout,
		},
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
		fmt.Println("Usage: rlmgo validate <file>")
		os.Exit(1)
	}

	filename := os.Args[2]
	src, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	_, err = rlmgo.Compile(filename, string(src))
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Validation successful")
}
