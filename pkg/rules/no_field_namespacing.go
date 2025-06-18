package rules

import (
	"fmt"
	"strings"

	"github.com/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// NoFieldNamespacing checks that fields don't repeat their parent type name
type NoFieldNamespacing struct{}

// NewNoFieldNamespacing creates a new instance of the NoFieldNamespacing rule
func NewNoFieldNamespacing() *NoFieldNamespacing {
	return &NoFieldNamespacing{}
}

// Name returns the rule name
func (r *NoFieldNamespacing) Name() string {
	return "no-field-namespacing"
}

// Description returns what this rule checks
func (r *NoFieldNamespacing) Description() string {
	return "Fields don't need to be namespaced with their parent type name - following Yelp guidelines"
}

// Check validates that fields don't unnecessarily repeat their parent type name
func (r *NoFieldNamespacing) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check fields in object types
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			for _, field := range def.Fields {
				// Skip introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}

				// Check if field name contains the type name unnecessarily
				if r.isFieldNamespaced(def.Name, field.Name) {
					line, column := 1, 1
					if field.Position != nil {
						line = field.Position.Line
						column = field.Position.Column
					}

					// Suggest a better name
					betterName := r.suggestBetterFieldName(def.Name, field.Name)

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Field `%s.%s` unnecessarily repeats the type name. Consider `%s` instead.", def.Name, field.Name, betterName),
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

// isFieldNamespaced checks if a field name unnecessarily includes the type name
func (r *NoFieldNamespacing) isFieldNamespaced(typeName, fieldName string) bool {
	// Convert type name to various common patterns
	lowerTypeName := strings.ToLower(typeName)
	camelTypeName := r.toCamelCase(typeName)

	// Common namespacing patterns to detect
	patterns := []string{
		lowerTypeName,
		camelTypeName,
		strings.ToLower(typeName) + "_",
	}

	lowerFieldName := strings.ToLower(fieldName)

	for _, pattern := range patterns {
		// Check if field starts with the type name pattern
		if strings.HasPrefix(lowerFieldName, pattern) && len(fieldName) > len(pattern) {
			// Make sure it's not just a coincidence (field should have more content after the prefix)
			remainder := fieldName[len(pattern):]
			if len(remainder) > 0 && (remainder[0] >= 'A' && remainder[0] <= 'Z') {
				return true
			}
		}
	}

	return false
}

// suggestBetterFieldName suggests a field name without the type prefix
func (r *NoFieldNamespacing) suggestBetterFieldName(typeName, fieldName string) string {
	// Convert type name to various common patterns
	lowerTypeName := strings.ToLower(typeName)
	camelTypeName := r.toCamelCase(typeName)

	patterns := []string{
		camelTypeName,
		lowerTypeName,
		strings.ToLower(typeName) + "_",
	}

	lowerFieldName := strings.ToLower(fieldName)

	for _, pattern := range patterns {
		if strings.HasPrefix(lowerFieldName, pattern) {
			remainder := fieldName[len(pattern):]
			if len(remainder) > 0 {
				// Make the first character lowercase for camelCase
				return strings.ToLower(remainder[:1]) + remainder[1:]
			}
		}
	}

	return fieldName
}

// toCamelCase converts PascalCase to camelCase
func (r *NoFieldNamespacing) toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
