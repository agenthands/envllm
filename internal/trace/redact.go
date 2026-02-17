package trace

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Redactor handles sensitive data masking in trace logs.
type Redactor struct {
	Policy string // "STRICT" or "DEBUG"
}

func NewRedactor(policy string) *Redactor {
	if policy == "" {
		policy = "STRICT"
	}
	return &Redactor{Policy: policy}
}

// Redact returns a masked version of the input if it contains sensitive patterns.
func (r *Redactor) Redact(input interface{}) interface{} {
	if r.Policy == "DEBUG" {
		return input
	}

	switch v := input.(type) {
	case string:
		return r.redactString(v)
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for k, val := range v {
			if isSensitiveKey(k) {
				newMap[k] = r.redactString(fmt.Sprintf("%v", val))
			} else {
				newMap[k] = r.Redact(val)
			}
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(v))
		for i, val := range v {
			newSlice[i] = r.Redact(val)
		}
		return newSlice
	default:
		return v
	}
}

func (r *Redactor) redactString(s string) string {
	if len(s) == 0 {
		return ""
	}
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("[REDACTED:len=%d,hash=%x]", len(s), h[:4])
}

func isSensitiveKey(k string) bool {
	k = strings.ToLower(k)
	return strings.Contains(k, "pass") || 
		   strings.Contains(k, "token") || 
		   strings.Contains(k, "key") || 
		   strings.Contains(k, "secret") || 
		   strings.Contains(k, "auth")
}
