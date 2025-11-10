package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// MutationResponseNullable checks that mutation root fields are nullable
type MutationResponseNullable struct{}

// NewMutationResponseNullable creates a new instance of the MutationResponseNullable rule
func NewMutationResponseNullable() *MutationResponseNullable {
	return &MutationResponseNullable{}
}

// Name returns the rule name
func (r *MutationResponseNullable) Name() string {
	return "mutation-response-nullable"
}

// Description returns what this rule checks
func (r *MutationResponseNullable) Description() string {
	return "Mutation root fields should be nullable to prevent the entire mutation response from becoming null when data is missing."
}

// Check validates that mutation response fields are nullable
func (r *MutationResponseNullable) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find the Mutation type
	mutationType := schema.Types["Mutation"]
	if mutationType == nil {
		return errors // No mutations to check
	}

	for _, field := range mutationType.Fields {
		// Skip introspection types
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		// Enforce that the top-level field is nullable. Inner/nested types are not checked.
		if field.Type != nil && field.Type.NonNull {
			line, column := 1, 1
			if field.Position != nil {
				line = field.Position.Line
				column = field.Position.Column
			}

			suggestion := r.makeNullable(field.Type)

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Mutation root field `%s` should be nullable (`%s` instead of `%s`) to prevent the entire mutation response from becoming null when data is missing.",
					field.Name, suggestion, field.Type.String(),
				),
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

// makeNullable converts a non-null type to nullable
func (r *MutationResponseNullable) makeNullable(fieldType *ast.Type) string {
	typeStr := fieldType.String()
	// Remove the ! at the end if present
	if strings.HasSuffix(typeStr, "!") {
		return typeStr[:len(typeStr)-1]
	}

	return typeStr
}
