package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// Alphabetize checks that fields and enum values are in alphabetical order
type Alphabetize struct{}

// NewAlphabetize creates a new instance of the Alphabetize rule
func NewAlphabetize() *Alphabetize {
	return &Alphabetize{}
}

// Name returns the rule name
func (r *Alphabetize) Name() string {
	return "alphabetize"
}

// Description returns what this rule checks
func (r *Alphabetize) Description() string {
	return "Enforce alphabetical order for type fields and enum values - following Guild best practices for consistency"
}

// Check validates that fields and enum values are alphabetically ordered
func (r *Alphabetize) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check fields in object types and interfaces
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			// Get field names, filtering out built-in and introspection fields
			var fieldNames []string
			for _, field := range def.Fields {
				// Skip built-in fields and introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}
				fieldNames = append(fieldNames, field.Name)
			}

			// Check if fields are alphabetically ordered
			if len(fieldNames) > 1 && !r.isAlphabeticallyOrdered(fieldNames) {
				line, column := 1, 1
				if def.Position != nil {
					line = def.Position.Line
					column = def.Position.Column
				}

				sortedNames := make([]string, len(fieldNames))
				copy(sortedNames, fieldNames)
				sort.Strings(sortedNames)

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Fields in type `%s` should be alphabetically ordered. Expected order: [%s]", def.Name, strings.Join(sortedNames, ", ")),
					Location: types.Location{
						Line:   line,
						Column: column,
						File:   source.Name,
					},
					Rule: r.Name(),
				})
			}
		}

		// Check enum values
		if def.Kind == ast.Enum {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			// Get enum value names
			enumNames := make([]string, len(def.EnumValues))
			for i, enumValue := range def.EnumValues {
				enumNames[i] = enumValue.Name
			}

			// Check if enum values are alphabetically ordered
			if !r.isAlphabeticallyOrdered(enumNames) {
				line, column := 1, 1
				if def.Position != nil {
					line = def.Position.Line
					column = def.Position.Column
				}

				sortedNames := make([]string, len(enumNames))
				copy(sortedNames, enumNames)
				sort.Strings(sortedNames)

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Enum values in `%s` should be alphabetically ordered. Expected order: [%s]", def.Name, strings.Join(sortedNames, ", ")),
					Location: types.Location{
						Line:   line,
						Column: column,
						File:   source.Name,
					},
					Rule: r.Name(),
				})
			}
		}

		// Check input object fields
		if def.Kind == ast.InputObject {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			// Get field names, filtering out built-in and introspection fields
			var fieldNames []string
			for _, field := range def.Fields {
				// Skip built-in fields and introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}
				fieldNames = append(fieldNames, field.Name)
			}

			// Check if fields are alphabetically ordered
			if len(fieldNames) > 1 && !r.isAlphabeticallyOrdered(fieldNames) {
				line, column := 1, 1
				if def.Position != nil {
					line = def.Position.Line
					column = def.Position.Column
				}

				sortedNames := make([]string, len(fieldNames))
				copy(sortedNames, fieldNames)
				sort.Strings(sortedNames)

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Fields in input type `%s` should be alphabetically ordered. Expected order: [%s]", def.Name, strings.Join(sortedNames, ", ")),
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

// isAlphabeticallyOrdered checks if a slice of strings is alphabetically ordered
func (r *Alphabetize) isAlphabeticallyOrdered(names []string) bool {
	if len(names) <= 1 {
		return true
	}

	for i := 1; i < len(names); i++ {
		// Case-insensitive comparison
		if strings.ToLower(names[i-1]) > strings.ToLower(names[i]) {
			return false
		}
	}

	return true
}
