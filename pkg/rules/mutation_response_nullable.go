package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// MutationResponseNullable checks that mutation response fields are nullable
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
	return "Mutation response fields should be nullable to prevent breaking changes during schema evolution"
}

// Check validates that mutation response fields are nullable
func (r *MutationResponseNullable) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find the Mutation type
	mutationType := schema.Types["Mutation"]
	if mutationType == nil {
		return errors // No mutations to check
	}

	// Collect all mutation return types
	mutationReturnTypes := make(map[string]bool)
	for _, field := range mutationType.Fields {
		typeName := r.getTypeName(field.Type)
		if typeName != "" {
			mutationReturnTypes[typeName] = true
		}
	}

	// Check each mutation return type
	for typeName := range mutationReturnTypes {
		responseType := schema.Types[typeName]
		if responseType == nil || responseType.Kind != ast.Object {
			continue
		}

		// Skip introspection types
		if strings.HasPrefix(typeName, "__") {
			continue
		}

		// Check all fields in the response type
		for _, field := range responseType.Fields {
			// Skip introspection fields
			if strings.HasPrefix(field.Name, "__") {
				continue
			}

			if r.isNonNullType(field.Type) {
				line, column := 1, 1
				if field.Position != nil {
					line = field.Position.Line
					column = field.Position.Column
				}

				suggestion := r.makeNullable(field.Type)

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Mutation response field `%s.%s` should be nullable (`%s` instead of `%s`) to prevent breaking changes when evolving the schema.", typeName, field.Name, suggestion, r.typeToString(field.Type)),
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

// getTypeName extracts the type name from a potentially wrapped type
func (r *MutationResponseNullable) getTypeName(fieldType *ast.Type) string {
	current := fieldType

	// Unwrap all wrappers to get to the named type
	for current != nil {
		if current.NamedType != "" {
			return current.NamedType
		}
		current = current.Elem
	}

	return ""
}

// isNonNullType checks if a type is non-null
func (r *MutationResponseNullable) isNonNullType(fieldType *ast.Type) bool {
	return fieldType.NonNull && fieldType.NamedType != ""
}

// makeNullable converts a non-null type to nullable
func (r *MutationResponseNullable) makeNullable(fieldType *ast.Type) string {
	typeStr := r.typeToString(fieldType)

	// Remove the ! at the end if present
	if strings.HasSuffix(typeStr, "!") {
		return typeStr[:len(typeStr)-1]
	}

	return typeStr
}

// typeToString converts an AST type to its string representation
func (r *MutationResponseNullable) typeToString(fieldType *ast.Type) string {
	if fieldType.NamedType != "" {
		if fieldType.NonNull {
			return fieldType.NamedType + "!"
		}
		return fieldType.NamedType
	}

	if fieldType.Elem != nil {
		innerStr := r.typeToString(fieldType.Elem)
		if fieldType.NonNull {
			return "[" + innerStr + "]!"
		}
		return "[" + innerStr + "]"
	}

	return "Unknown"
}
