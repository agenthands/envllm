package capability

import (
	"fmt"
	"os"

	"github.com/agenthands/rlm-go/internal/runtime"
)

// ReadFile implements the READ_FILE operation.
func ReadFile(s *runtime.Session, path runtime.Value) (runtime.Value, error) {
	p := ""
	if path.Kind == runtime.KindText {
		p, _ = s.Stores.Text.Get(path.V.(runtime.TextHandle))
	} else {
		p = path.V.(string)
	}

	if err := s.ValidatePath(p, false); err != nil {
		return runtime.Value{}, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return runtime.Value{}, fmt.Errorf("READ_FILE failed: %v", err)
	}

	h := s.Stores.Text.Add(string(data))
	return runtime.Value{Kind: runtime.KindText, V: h}, nil
}

// WriteFile implements the WRITE_FILE operation.
func WriteFile(s *runtime.Session, path runtime.Value, source runtime.Value) (runtime.Value, error) {
	p := ""
	if path.Kind == runtime.KindText {
		p, _ = s.Stores.Text.Get(path.V.(runtime.TextHandle))
	} else {
		p = path.V.(string)
	}

	if err := s.ValidatePath(p, true); err != nil {
		return runtime.Value{}, err
	}

	text := ""
	if source.Kind == runtime.KindText {
		text, _ = s.Stores.Text.Get(source.V.(runtime.TextHandle))
	} else {
		text = source.V.(string)
	}

	if err := os.WriteFile(p, []byte(text), 0644); err != nil {
		return runtime.Value{}, fmt.Errorf("WRITE_FILE failed: %v", err)
	}

	return runtime.Value{Kind: runtime.KindBool, V: true}, nil
}

// ListDir implements the LIST_DIR operation.
func ListDir(s *runtime.Session, path runtime.Value) (runtime.Value, error) {
	p := ""
	if path.Kind == runtime.KindText {
		p, _ = s.Stores.Text.Get(path.V.(runtime.TextHandle))
	} else {
		p = path.V.(string)
	}

	if err := s.ValidatePath(p, false); err != nil {
		return runtime.Value{}, err
	}

	entries, err := os.ReadDir(p)
	if err != nil {
		return runtime.Value{}, fmt.Errorf("LIST_DIR failed: %v", err)
	}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return runtime.Value{Kind: runtime.KindJSON, V: names}, nil
}
