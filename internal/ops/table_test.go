package ops

import (
	"testing"
)

func TestLoadTable(t *testing.T) {
	tbl, err := LoadTable("../../assets/ops.json")
	if err != nil {
		t.Fatalf("LoadTable failed: %v", err)
	}

	if tbl.Version != "0.1" {
		t.Errorf("expected version 0.1, got %s", tbl.Version)
	}

	if _, ok := tbl.Ops["STATS"]; !ok {
		t.Errorf("STATS op not found")
	}

	if _, ok := tbl.Ops["FIND_TEXT"]; !ok {
		t.Errorf("FIND_TEXT op not found")
	}
}
