package capability

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agenthands/envllm/internal/runtime"
)

type mockTextStore struct {
	content map[string]string
	nextID int
}

func (m *mockTextStore) Add(text string) runtime.TextHandle {
	m.nextID++
	id := filepath.Base(text) // Just for test
	if len(id) > 10 { id = "t1" }
	m.content[id] = text
	return runtime.TextHandle{ID: id, Bytes: len(text)}
}
func (m *mockTextStore) Get(h runtime.TextHandle) (string, bool) {
	t, ok := m.content[h.ID]
	return t, ok
}
func (m *mockTextStore) Window(h runtime.TextHandle, center, radius int) (runtime.TextHandle, error) {
	return runtime.TextHandle{}, nil
}

func TestFS_Ops(t *testing.T) {
	tmpDir := t.TempDir()
	readDir := filepath.Join(tmpDir, "read")
	writeDir := filepath.Join(tmpDir, "write")
	os.Mkdir(readDir, 0755)
	os.Mkdir(writeDir, 0755)

	policy := runtime.Policy{
		AllowedReadPaths:  []string{readDir, writeDir},
		AllowedWritePaths: []string{writeDir},
	}
	ts := &mockTextStore{content: make(map[string]string)}
	s := runtime.NewSession(policy, ts)

	// Test WRITE_FILE
	filePath := filepath.Join(writeDir, "test.txt")
	_, err := WriteFile(s, runtime.Value{Kind: runtime.KindString, V: filePath}, runtime.Value{Kind: runtime.KindString, V: "hello"})
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Test READ_FILE
	res, err := ReadFile(s, runtime.Value{Kind: runtime.KindString, V: filePath})
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	h := res.V.(runtime.TextHandle)
	content, _ := ts.Get(h)
	if content != "hello" {
		t.Errorf("expected 'hello', got %q", content)
	}

	// Test LIST_DIR
	res, err = ListDir(s, runtime.Value{Kind: runtime.KindString, V: writeDir})
	if err != nil {
		t.Fatalf("ListDir failed: %v", err)
	}
	list := res.V.([]string)
	if len(list) != 1 || list[0] != "test.txt" {
		t.Errorf("expected ['test.txt'], got %v", list)
	}

	// Test KindText path
	pathHandle := ts.Add(filePath)
	res, err = ReadFile(s, runtime.Value{Kind: runtime.KindText, V: pathHandle})
	if err != nil {
		t.Fatalf("ReadFile with KindText path failed: %v", err)
	}

	// Test Errors
	_, err = ReadFile(s, runtime.Value{Kind: runtime.KindString, V: filepath.Join(readDir, "non-existent")})
	if err == nil {
		t.Errorf("expected error for non-existent file")
	}

	_, err = ListDir(s, runtime.Value{Kind: runtime.KindString, V: filepath.Join(readDir, "non-existent")})
	if err == nil {
		t.Errorf("expected error for non-existent dir")
	}
	
	// Test WriteFile with KindText source
	srcHandle := ts.Add("content from handle")
	_, err = WriteFile(s, runtime.Value{Kind: runtime.KindString, V: filePath}, runtime.Value{Kind: runtime.KindText, V: srcHandle})
	if err != nil {
		t.Fatalf("WriteFile with KindText source failed: %v", err)
	}

	// Test security_error
	secretFile := filepath.Join(tmpDir, "secret.txt")
	os.WriteFile(secretFile, []byte("data"), 0644)
	_, err = ReadFile(s, runtime.Value{Kind: runtime.KindString, V: secretFile})
	if err == nil {
		t.Errorf("expected security error for non-whitelisted path")
	}

	_, err = WriteFile(s, runtime.Value{Kind: runtime.KindString, V: secretFile}, runtime.Value{Kind: runtime.KindString, V: "bad"})
	if err == nil {
		t.Errorf("expected security error for WriteFile")
	}

	// Test WriteFile OS error (e.g. writing to a directory)
	_, err = WriteFile(s, runtime.Value{Kind: runtime.KindString, V: writeDir}, runtime.Value{Kind: runtime.KindString, V: "content"})
	if err == nil {
		t.Errorf("expected OS error for WriteFile to directory")
	}
}
