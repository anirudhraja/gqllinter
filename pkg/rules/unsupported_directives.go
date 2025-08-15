package rules

import (
	"fmt"
	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// UnsupportedDirectives checks that no unsupported directives are used
type UnsupportedDirectives struct{}

// NewUnsupportedDirectives creates a new instance of the UnsupportedDirectives rule
func NewUnsupportedDirectives() *UnsupportedDirectives {
	return &UnsupportedDirectives{}
}

// Name returns the rule name
func (r *UnsupportedDirectives) Name() string {
	return "unsupported-directives"
}

// Description returns what this rule checks
func (r *UnsupportedDirectives) Description() string {
	return "No unsupported directives should be used in the schemas"
}

// Check validates no unsupported directives are used
func (r *UnsupportedDirectives) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Map of supported directives
	supportedDirectivesMap := map[string]bool{
		"link":          true,
		"key":           true,
		"shareable":     true,
		"external":      true,
		"error":         true,
		"throws":        true,
		"responseUnion": true,
		"include":       true,
		"skip":          true,
		"deprecated":    true,
		"specifiedBy":   true,
		"defer":         true,
		"oneOf":         true,
	}

	for _, dir := range schema.Directives {
		if !supportedDirectivesMap[dir.Name] {
			line, column := 1, 1
			if dir.Position != nil {
				line = dir.Position.Line
				column = dir.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("The schema uses unsupported directive @%s. This directive is not supported in this schema.", dir.Name),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}
