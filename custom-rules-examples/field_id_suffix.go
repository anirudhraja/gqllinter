package main

import (
	"fmt"
	"strings"

	"github.com/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// FieldIdSuffixRule checks that ID fields end with 'ID' not 'Id'
type FieldIdSuffixRule struct{}

// NewRule is the required entry point for plugins
func NewRule() types.Rule {
	return &FieldIdSuffixRule{}
}

// Name returns the rule identifier
func (r *FieldIdSuffixRule) Name() string {
	return "field-id-suffix"
}

// Description explains what this rule does
func (r *FieldIdSuffixRule) Description() string {
	return "Ensures ID fields end with 'ID' not 'Id' for consistency"
}

// Check performs the actual linting
func (r *FieldIdSuffixRule) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	for _, def := range schema.Types {
		// Skip introspection types
		if strings.HasPrefix(def.Name, "__") {
			continue
		}

		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.InputObject {
			for _, field := range def.Fields {
				// Skip introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}

				// Check if field ends with 'Id' but not 'ID'
				if strings.HasSuffix(field.Name, "Id") && !strings.HasSuffix(field.Name, "ID") {
					line, column := 1, 1
					if field.Position != nil {
						line = field.Position.Line
						column = field.Position.Column
					}

					suggestedName := strings.TrimSuffix(field.Name, "Id") + "ID"

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Field `%s.%s` should end with 'ID' not 'Id'. Consider renaming to `%s`.", def.Name, field.Name, suggestedName),
						Location: types.Location{
							Line:   line,
							Column: column,
							File:   source.Name,
						},
						Rule: r.Name(),
					})
				}
			}
		}
	}

	return errors
}
