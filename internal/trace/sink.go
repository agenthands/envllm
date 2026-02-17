package trace

import (
	"encoding/json"
	"os"
	"sync"
)

type Sink interface {
	Emit(step TraceStep) error
	Close() error
}

type JSONLSink struct {
	mu   sync.Mutex
	file *os.File
}

func NewJSONLSink(path string) (*JSONLSink, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &JSONLSink{file: f}, nil
}

func (s *JSONLSink) Emit(step TraceStep) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if step.TraceVersion == "" {
		step.TraceVersion = "1"
	}
	
	data, err := json.Marshal(step)
	if err != nil {
		return err
	}
	
	_, err = s.file.Write(data)
	if err != nil {
		return err
	}
	_, err = s.file.Write([]byte("\n"))
	return err
}

func (s *JSONLSink) Close() error {
	return s.file.Close()
}

// MemorySink for testing
type MemorySink struct {
	mu    sync.Mutex
	Steps []TraceStep
}

func (s *MemorySink) Emit(step TraceStep) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Steps = append(s.Steps, step)
	return nil
}

func (s *MemorySink) Close() error {
	return nil
}
