package extension

import (
	"context"
	"fmt"

	"github.com/agenthands/envllm/internal/ast"
	"github.com/agenthands/envllm/internal/rewrite"
)

// MappingEngine handles version negotiation and script upgrading.
type MappingEngine struct {
	Manifests     map[string]Manifest
	RewriteEngine *rewrite.Engine
}

func NewMappingEngine(rw *rewrite.Engine) *MappingEngine {
	return &MappingEngine{
		Manifests:     make(map[string]Manifest),
		RewriteEngine: rw,
	}
}

func (e *MappingEngine) Register(m Manifest) {
	e.Manifests[m.Name] = m
}

// UpgradeProgram applies relevant mappings to the program AST.
func (e *MappingEngine) UpgradeProgram(ctx context.Context, prog *ast.Program) ([]string, error) {
	appliedRules := []string{}

	// 1. Negotiation: Check if requested versions are supported
	for ext, reqVer := range prog.Extensions {
		m, ok := e.Manifests[ext]
		if !ok {
			return nil, fmt.Errorf("ERR_EXTENSION_NOT_FOUND: %s", ext)
		}
		
		// Simple equality check for now, can be semver range later
		if m.Version != reqVer {
			// Find mappings for this extension range
			for _, mapping := range m.Compat.Mappings {
				// Match on extension version
				if mapping.ExtensionRange == reqVer || mapping.ExtensionRange == "*" {
					// Apply the mapping/rewrite rule
					if e.RewriteEngine != nil {
						var err error
						prog, err = e.RewriteEngine.ApplyRules(ctx, prog, []string{mapping.RewriteRuleID})
						if err != nil {
							return appliedRules, fmt.Errorf("mapping %s failed: %v", mapping.ID, err)
						}
					}
					appliedRules = append(appliedRules, mapping.RewriteRuleID)
				}
			}
		}
	}

	return appliedRules, nil
}
