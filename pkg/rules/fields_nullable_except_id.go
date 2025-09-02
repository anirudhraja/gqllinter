package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// FieldsNullableExceptId checks that all fields are nullable except ID fields
type FieldsNullableExceptId struct {
	// excludedTypes contains object/type names that should be excluded from this rule
	excludedTypes map[string]bool
}

// NewFieldsNullableExceptId creates a new instance of the FieldsNullableExceptId rule
func NewFieldsNullableExceptId() *FieldsNullableExceptId {
	excludedTypes := map[string]bool{
		"PageInfo": true,
	}

	return &FieldsNullableExceptId{
		excludedTypes: excludedTypes,
	}
}

// Name returns the rule name
func (r *FieldsNullableExceptId) Name() string {
	return "fields-nullable-except-id"
}

// Description returns what this rule checks
func (r *FieldsNullableExceptId) Description() string {
	return "All fields should be nullable except ID fields to enable better schema evolution and avoid breaking changes"
}

// Check validates that all fields are nullable except ID fields
func (r *FieldsNullableExceptId) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all object types
	for _, def := range schema.Types {
		// Skip introspection types, root types, and excluded types
		if strings.HasPrefix(def.Name, "__") {
			continue
		}
		if def.Kind == ast.Object {
			if def.Name == "Query" ||
				def.Name == "Mutation" ||
				def.Name == "Subscription" ||
				r.excludedTypes[def.Name] ||
				r.isConnectionType(def.Name) ||
				r.isEdgeType(def.Name) {
				continue
			}

			// Check each field in the type
			for _, field := range def.Fields {
				// Skip introspection types, root types, and excluded types
				if strings.HasPrefix(field.Name, "__") {
					continue
				}
				if r.shouldBeNullable(field) && r.isNonNullType(field.Type) {
					line, column := 1, 1
					if field.Position != nil {
						line = field.Position.Line
						column = field.Position.Column
					}

					suggestion := r.makeNullable(field.Type)
					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Field `%s.%s` should be nullable (`%s` instead of `%s`) to enable schema evolution and avoid breaking changes.",
							def.Name, field.Name, suggestion, r.typeToString(field.Type)),
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

// shouldBeNullable determines if a field should be nullable (all except ID fields)
func (r *FieldsNullableExceptId) shouldBeNullable(field *ast.FieldDefinition) bool {
	// ID fields can be non-null
	if field.Name == "id" || strings.HasSuffix(field.Name, "Id") || strings.HasSuffix(field.Name, "ID") {
		// Additional check: make sure it's actually an ID type
		typeName := r.getTypeName(field.Type)
		if typeName == "ID" {
			return false
		}
	}

	// All other fields should be nullable
	return true
}

// getTypeName extracts the base type name from a field type
func (r *FieldsNullableExceptId) getTypeName(fieldType *ast.Type) string {
	if fieldType.NamedType != "" {
		return fieldType.NamedType
	}
	if fieldType.Elem != nil {
		return r.getTypeName(fieldType.Elem)
	}
	return ""
}

// isNonNullType checks if a type is non-null
func (r *FieldsNullableExceptId) isNonNullType(fieldType *ast.Type) bool {
	return fieldType.NonNull
}

// makeNullable converts a non-null type to nullable
func (r *FieldsNullableExceptId) makeNullable(fieldType *ast.Type) string {
	if fieldType.NonNull {
		// Remove the NonNull wrapper
		if fieldType.NamedType != "" {
			return fieldType.NamedType
		}
		if fieldType.Elem != nil {
			return r.typeToString(fieldType.Elem)
		}
	}
	return r.typeToString(fieldType)
}

// typeToString converts a type to its string representation
func (r *FieldsNullableExceptId) typeToString(fieldType *ast.Type) string {
	if fieldType.NonNull {
		if fieldType.NamedType != "" {
			return fieldType.NamedType + "!"
		}
		if fieldType.Elem != nil {
			return r.typeToString(fieldType.Elem) + "!"
		}
	}

	if fieldType.NamedType != "" {
		return fieldType.NamedType
	}

	if fieldType.Elem != nil {
		return "[" + r.typeToString(fieldType.Elem) + "]"
	}

	return "Unknown"
}

func (r *FieldsNullableExceptId) isConnectionType(typeName string) bool {
	return strings.HasSuffix(strings.ToLower(typeName), "connection")
}

func (r *FieldsNullableExceptId) isEdgeType(typeName string) bool {
	return strings.HasSuffix(strings.ToLower(typeName), "edge")
}
