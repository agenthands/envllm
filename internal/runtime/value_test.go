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
