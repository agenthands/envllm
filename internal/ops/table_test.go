package ops

import (
	"testing"
	"github.com/agenthands/envllm/internal/runtime"
)

func TestLoadTable(t *testing.T) {
	tbl, err := LoadTable("../../assets/ops.json")
	if err != nil {
		t.Fatalf("LoadTable failed: %v", err)
	}

	if tbl.Version != "0.2" {
		t.Errorf("expected version 0.2, got %s", tbl.Version)
	}

	if _, ok := tbl.Ops["STATS"]; !ok {
		t.Errorf("STATS op not found")
	}

	if _, ok := tbl.Ops["FIND_TEXT"]; !ok {
		t.Errorf("FIND_TEXT op not found")
	}
}

func TestTable_ValidateSignature(t *testing.T) {
	tbl, _ := LoadTable("../../assets/ops.json")

	tests := []struct {
		name    string
		opName  string
		args    []ValidatedKwArg
		wantErr bool
	}{
		{
			name:   "Valid STATS",
			opName: "STATS",
			args: []ValidatedKwArg{
				{Keyword: "SOURCE", Value: runtime.Value{Kind: runtime.KindText}},
			},
			wantErr: false,
		},
		{
			name:   "Wrong Keyword",
			opName: "STATS",
			args: []ValidatedKwArg{
				{Keyword: "SRC", Value: runtime.Value{Kind: runtime.KindText}},
			},
			wantErr: true,
		},
		{
			name:   "Wrong Type",
			opName: "STATS",
			args: []ValidatedKwArg{
				{Keyword: "SOURCE", Value: runtime.Value{Kind: runtime.KindInt}},
			},
			wantErr: true,
		},
		{
			name:   "Enum Valid",
			opName: "FIND_TEXT",
			args: []ValidatedKwArg{
				{Keyword: "SOURCE", Value: runtime.Value{Kind: runtime.KindText}},
				{Keyword: "NEEDLE", Value: runtime.Value{Kind: runtime.KindText}},
				{Keyword: "MODE", Value: runtime.Value{Kind: runtime.KindJSON, V: "FIRST"}},
				{Keyword: "IGNORE_CASE", Value: runtime.Value{Kind: runtime.KindBool}},
			},
			wantErr: false,
		},
		{
			name:   "Enum Invalid",
			opName: "FIND_TEXT",
			args: []ValidatedKwArg{
				{Keyword: "SOURCE", Value: runtime.Value{Kind: runtime.KindText}},
				{Keyword: "NEEDLE", Value: runtime.Value{Kind: runtime.KindText}},
				{Keyword: "MODE", Value: runtime.Value{Kind: runtime.KindJSON, V: "ALL"}},
				{Keyword: "IGNORE_CASE", Value: runtime.Value{Kind: runtime.KindBool}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tbl.ValidateSignature(tt.opName, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
