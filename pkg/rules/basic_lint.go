package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// BasicLint implements basic naming convention and type validation rules
type BasicLint struct{}

// NewBasicLint creates a new instance of BasicLint
func NewBasicLint() *BasicLint {
	return &BasicLint{}
}

// Name returns the name of this rule
func (r *BasicLint) Name() string {
	return "basic-lint"
}

// Description returns the description of this rule
func (r *BasicLint) Description() string {
	return "Validates basic naming conventions: @error types should not have 'Error' suffix (proto translation adds ErrorDetails), and no type should contain 'PropertiesBy' substring"
}

// Check performs the basic lint validation
func (r *BasicLint) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all types in the schema
	for typeName, typeDefinition := range schema.Types {
		// Skip built-in types
		if typeDefinition.BuiltIn {
			continue
		}

		// Rule 1: @error types shouldn't have 'Error' trailing in their names
		if r.hasErrorDirective(typeDefinition) {
			if r.hasErrorSuffix(typeName) {
				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Type '%s' has @error directive but contains 'Error' in its name. Remove the 'Error' suffix", typeName),
					Location: types.Location{
						Line:   typeDefinition.Position.Line,
						Column: typeDefinition.Position.Column,
						File:   source.Name,
					},
					Rule: r.Name(),
				})
			}
		}

		// Rule 2: No type should contain 'PropertiesBy' substring
		if r.containsPropertiesBy(typeName) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Type '%s' contains 'PropertiesBy' substring which is not allowed in type names", typeName),
				Location: types.Location{
					Line:   typeDefinition.Position.Line,
					Column: typeDefinition.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}

// hasErrorDirective checks if a type has the @error directive
func (r *BasicLint) hasErrorDirective(typeDefinition *ast.Definition) bool {
	if typeDefinition == nil {
		return false
	}

	for _, directive := range typeDefinition.Directives {
		if directive.Name == "error" {
			return true
		}
	}
	return false
}

// hasErrorSuffix checks if a type name has 'Error' suffix (case-insensitive)
func (r *BasicLint) hasErrorSuffix(typeName string) bool {
	lowerName := strings.ToLower(typeName)
	return strings.HasSuffix(lowerName, "error")
}

// containsPropertiesBy checks if a type name contains 'PropertiesBy' substring (case-insensitive)
func (r *BasicLint) containsPropertiesBy(typeName string) bool {
	lowerName := strings.ToLower(typeName)
	return strings.Contains(lowerName, "propertiesby")
}
