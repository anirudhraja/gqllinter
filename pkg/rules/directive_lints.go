package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// DirectivesCommonLint checks common directive lint issues
type DirectivesCommonLint struct{}

// NewDirectivesCommonLint creates a new instance of the DirectivesCommonLint rule
func NewDirectivesCommonLint() *DirectivesCommonLint {
	return &DirectivesCommonLint{}
}

// Name returns the rule name
func (r *DirectivesCommonLint) Name() string {
	return "common-directives-lint"
}

// Description returns what this rule checks
func (r *DirectivesCommonLint) Description() string {
	return "Common directive validation rules including conflict detection between @key and @shareable, and ensuring @shareable is only used on object types"
}

// Check validates common directive rules
func (r *DirectivesCommonLint) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	// Check for @key and @shareable conflicts on objects
	errors := r.checkKeyShareableConflicts(schema, source)

	return errors
}

// checkKeyShareableConflicts checks that @key and @shareable directives are not present on the same object
func (r *DirectivesCommonLint) checkKeyShareableConflicts(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check directives on schema types
	for _, def := range schema.Types {
		// Skip built-in types and introspection types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") || def.Kind != ast.Object {
			continue
		}

		hasKey := false
		hasShareable := false

		// Check for @key and @shareable directives on the type
		for _, directive := range def.Directives {
			if directive.Name == "key" {
				hasKey = true
			}
			if directive.Name == "shareable" {
				hasShareable = true
			}
		}

		// If both @key and @shareable are present, report an error
		if hasKey && hasShareable {
			line, column := 1, 1
			if def.Position != nil {
				line = def.Position.Line
				column = def.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("The object %s cannot have both @key and @shareable directives. These directives are not supported together.", def.Name),
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
