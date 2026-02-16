package metrics

import (
	"encoding/json"
	"os"
	"sync"
)

type SessionMetrics struct {
	mu sync.Mutex

	ParseAttempts      int `json:"parse_attempts"`
	ParseSuccesses     int `json:"parse_successes"`
	LintAttempts       int `json:"lint_attempts"`
	LintSuccesses      int `json:"lint_successes"`
	RepairIterations   int `json:"repair_iterations"`
	DriftViolations    int `json:"drift_violations"`
}

func (m *SessionMetrics) RecordParse(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ParseAttempts++
	if success {
		m.ParseSuccesses++
	}
}

func (m *SessionMetrics) RecordLint(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LintAttempts++
	if success {
		m.LintSuccesses++
	}
}

func (m *SessionMetrics) RecordRepair() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RepairIterations++
}

func (m *SessionMetrics) SaveJSON(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
