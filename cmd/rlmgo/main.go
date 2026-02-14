package main

import (
	"flag"
	"fmt"
	"os"
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
		repl()
	case "validate":
		validate()
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Printf("Unknown command: %s
", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage: rlmgo <command> [arguments]")
	fmt.Println("
Commands:")
	fmt.Println("  run <file>      Execute an RLMDSL script")
	fmt.Println("  repl            Start an interactive REPL")
	fmt.Println("  validate <file> Validate script syntax and ops")
	fmt.Println("  help            Show this help text")
}

func run() {
	fmt.Println("Command 'run' not yet implemented")
}

func repl() {
	fmt.Println("Command 'repl' not yet implemented")
}

func validate() {
	fmt.Println("Command 'validate' not yet implemented")
}
