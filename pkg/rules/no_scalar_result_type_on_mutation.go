package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// NoScalarResultTypeOnMutation checks that mutations don't return scalar types
type NoScalarResultTypeOnMutation struct{}

// NewNoScalarResultTypeOnMutation creates a new instance of the NoScalarResultTypeOnMutation rule
func NewNoScalarResultTypeOnMutation() *NoScalarResultTypeOnMutation {
	return &NoScalarResultTypeOnMutation{}
}

// Name returns the rule name
func (r *NoScalarResultTypeOnMutation) Name() string {
	return "no-scalar-result-type-on-mutation"
}

// Description returns what this rule checks
func (r *NoScalarResultTypeOnMutation) Description() string {
	return "Mutations should return object types, not scalars - following Guild best practices for better error handling"
}

// Check validates that mutation fields return object types instead of scalars
func (r *NoScalarResultTypeOnMutation) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
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

		fieldType := field.Type
		baseTypeName := r.getBaseTypeName(fieldType)

		// Check if the return type is a scalar
		if r.isScalarType(baseTypeName) {
			line, column := 1, 1
			if field.Position != nil {
				line = field.Position.Line
				column = field.Position.Column
			}

			suggestedType := r.suggestObjectType(field.Name, baseTypeName)

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Mutation `%s` returns scalar type `%s`. Consider returning an object type like `%s` for better error handling and extensibility.", field.Name, baseTypeName, suggestedType),
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

// getBaseTypeName extracts the base type name from a field type
func (r *NoScalarResultTypeOnMutation) getBaseTypeName(fieldType *ast.Type) string {
	// Unwrap lists and non-nulls to get the base type
	baseType := fieldType
	for baseType.Elem != nil {
		baseType = baseType.Elem
	}
	return baseType.Name()
}

// isScalarType checks if a type name represents a scalar type
func (r *NoScalarResultTypeOnMutation) isScalarType(typeName string) bool {
	// Standard GraphQL scalar types
	scalarTypes := []string{
		"String", "Int", "Float", "Boolean", "ID",
	}

	for _, scalar := range scalarTypes {
		if typeName == scalar {
			return true
		}
	}

	// Could also be a custom scalar - in practice, most custom scalars follow naming patterns
	// This is a heuristic and could be made configurable
	return false
}

// suggestObjectType suggests an appropriate object type name for a mutation
func (r *NoScalarResultTypeOnMutation) suggestObjectType(mutationName, scalarType string) string {
	// Convert mutation name to a result type name
	mutationNameTrimmed := strings.TrimSuffix(mutationName, "Mutation")
	mutationNameTrimmed = strings.TrimSuffix(mutationNameTrimmed, "Command")

	// Capitalize first letter if not already
	if len(mutationNameTrimmed) > 0 {
		mutationNameTrimmed = strings.ToUpper(mutationNameTrimmed[:1]) + mutationNameTrimmed[1:]
	}

	// Suggest different patterns based on the mutation
	if strings.Contains(strings.ToLower(mutationName), "create") {
		return mutationNameTrimmed + "Payload"
	} else if strings.Contains(strings.ToLower(mutationName), "update") {
		return mutationNameTrimmed + "Payload"
	} else if strings.Contains(strings.ToLower(mutationName), "delete") {
		return mutationNameTrimmed + "Result"
	} else {
		return mutationNameTrimmed + "Payload"
	}
}
