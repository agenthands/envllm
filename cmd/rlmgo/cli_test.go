package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCLI_Validate(t *testing.T) {
	// Setup
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test validate
	os.Args = []string{"rlmgo", "validate", "../../test.rlm"}
	
	output := captureOutput(func() {
		main()
	})

	if !strings.Contains(output, "Validation successful") {
		t.Errorf("expected success message, got %q", output)
	}
}

func TestCLI_Run(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"rlmgo", "run", "../../test.rlm"}
	
	output := captureOutput(func() {
		// main calls os.Exit on error, but since we are in test it won't exit if we don't let it.
		// Actually, run() doesn't call Exit if it's successful.
		main()
	})

	if !strings.Contains(output, `"status":"ok"`) {
		t.Errorf("expected json with ok status, got %q", output)
	}
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
