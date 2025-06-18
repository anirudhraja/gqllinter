package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// InputName checks that mutation arguments follow the standard naming pattern
type InputName struct{}

// NewInputName creates a new instance of the InputName rule
func NewInputName() *InputName {
	return &InputName{}
}

// Name returns the rule name
func (r *InputName) Name() string {
	return "input-name"
}

// Description returns what this rule checks
func (r *InputName) Description() string {
	return "Require mutation argument to be always called 'input' and input type to be called Mutation name + 'Input' - following Guild best practices"
}

// Check validates that mutation arguments follow the standard naming pattern
func (r *InputName) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check if there's a Mutation type
	if schema.Mutation == nil {
		return errors
	}

	// Check each mutation field
	for _, field := range schema.Mutation.Fields {
		// Skip introspection fields
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		// Check mutation arguments
		if len(field.Arguments) > 0 {
			// For mutations with arguments, we expect a single "input" argument
			if len(field.Arguments) == 1 {
				arg := field.Arguments[0]

				// Check if the argument is named "input"
				if arg.Name != "input" {
					line, column := 1, 1
					if arg.Position != nil {
						line = arg.Position.Line
						column = arg.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Mutation `%s` argument should be named 'input', not '%s'.", field.Name, arg.Name),
						Location: types.Location{
							Line:   line,
							Column: column,
							File:   source.Name,
						},
						Rule: r.Name(),
					})
				}

				// Check if the input type follows the naming convention
				expectedInputType := r.capitalizeFirst(field.Name) + "Input"
				actualInputType := r.getBaseTypeName(arg.Type)

				if actualInputType != expectedInputType {
					line, column := 1, 1
					if arg.Position != nil {
						line = arg.Position.Line
						column = arg.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Mutation `%s` input type should be named `%s`, not `%s`.", field.Name, expectedInputType, actualInputType),
						Location: types.Location{
							Line:   line,
							Column: column,
							File:   source.Name,
						},
						Rule: r.Name(),
					})
				}
			} else if len(field.Arguments) > 1 {
				// Multiple arguments - suggest consolidating into an input type
				line, column := 1, 1
				if field.Position != nil {
					line = field.Position.Line
					column = field.Position.Column
				}

				expectedInputType := r.capitalizeFirst(field.Name) + "Input"
				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Mutation `%s` has %d arguments. Consider consolidating into a single 'input' argument of type `%s`.", field.Name, len(field.Arguments), expectedInputType),
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

	return errors
}

// getBaseTypeName extracts the base type name from a field type
func (r *InputName) getBaseTypeName(fieldType *ast.Type) string {
	// Unwrap lists and non-nulls to get the base type
	baseType := fieldType
	for baseType.Elem != nil {
		baseType = baseType.Elem
	}
	return baseType.Name()
}

// capitalizeFirst capitalizes the first letter of a string
func (r *InputName) capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
