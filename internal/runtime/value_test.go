package runtime

import (
	"encoding/json"
	"testing"
)

func TestValueJSON(t *testing.T) {
	tests := []struct {
		name string
		val  Value
		want string
	}{
		{
			name: "INT",
			val:  Value{Kind: KindInt, V: 450},
			want: `{"kind":"INT","v":450}`,
		},
		{
			name: "BOOL",
			val:  Value{Kind: KindBool, V: true},
			want: `{"kind":"BOOL","v":true}`,
		},
		{
			name: "SPAN",
			val:  Value{Kind: KindSpan, V: Span{Start: 10, End: 20}},
			want: `{"kind":"SPAN","v":{"start":10,"end":20}}`,
		},
		{
			name: "TEXT",
			val:  Value{Kind: KindText, V: TextHandle{ID: "t1", Bytes: 100}},
			want: `{"kind":"TEXT","v":{"id":"t1","bytes":100}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.val)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", string(got), tt.want)
			}

			var back Value
			if err := json.Unmarshal(got, &back); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if back.Kind != tt.val.Kind {
				t.Errorf("Unmarshal() Kind = %v, want %v", back.Kind, tt.val.Kind)
			}
		})
	}
}

func TestValueUnmarshal_JSON(t *testing.T) {
	data := `{"kind":"JSON","v":{"foo":"bar"}}`
	var v Value
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if v.Kind != KindJSON {
		t.Errorf("expected KindJSON")
	}
}

func TestValueUnmarshal_Span(t *testing.T) {
	data := `{"kind":"SPAN","v":{"start":10,"end":20}}`
	var v Value
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if v.Kind != KindSpan {
		t.Errorf("expected KindSpan")
	}
	s := v.V.(Span)
	if s.Start != 10 || s.End != 20 {
		t.Errorf("expected start 10, end 20, got %v", s)
	}
}

func TestValueUnmarshal_Text(t *testing.T) {
	data := `{"kind":"TEXT","v":{"id":"t1","bytes":100}}`
	var v Value
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if v.Kind != KindText {
		t.Errorf("expected KindText")
	}
}

func TestValueUnmarshal_Error(t *testing.T) {
	data := `{"kind":"UNKNOWN","v":1}`
	var v Value
	if err := json.Unmarshal([]byte(data), &v); err == nil {
		t.Errorf("expected error for unknown kind")
	}
}
