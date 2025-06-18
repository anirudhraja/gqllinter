package rules

import (
	"fmt"

	"github.com/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// FieldsHaveDescriptions checks that all fields have descriptions
type FieldsHaveDescriptions struct{}

// NewFieldsHaveDescriptions creates a new instance of the FieldsHaveDescriptions rule
func NewFieldsHaveDescriptions() *FieldsHaveDescriptions {
	return &FieldsHaveDescriptions{}
}

// Name returns the rule name
func (r *FieldsHaveDescriptions) Name() string {
	return "fields-have-descriptions"
}

// Description returns what this rule checks
func (r *FieldsHaveDescriptions) Description() string {
	return "All fields should have descriptions to explain their purpose"
}

// Check validates that all fields have descriptions
func (r *FieldsHaveDescriptions) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check fields in object types
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			for _, field := range def.Fields {
				if field.Description == "" {
					// For fields, position information might not be available in the schema built from source
					line, column := 1, 1
					if field.Position != nil {
						line = field.Position.Line
						column = field.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("The field `%s.%s` is missing a description.", def.Name, field.Name),
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
