package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// NamingConvention checks that types follow proper naming conventions
type NamingConvention struct{}

// NewNamingConvention creates a new instance of the NamingConvention rule
func NewNamingConvention() *NamingConvention {
	return &NamingConvention{}
}

// Name returns the rule name
func (r *NamingConvention) Name() string {
	return "naming-convention"
}

// Description returns what this rule checks
func (r *NamingConvention) Description() string {
	return "Enforce specific naming conventions - be specific with type names, avoid generic names"
}

// Check validates naming conventions
func (r *NamingConvention) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check for overly type names
	for _, def := range schema.Types {
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		line, column := 1, 1
		if def.Position != nil {
			line = def.Position.Line
			column = def.Position.Column
		}

		// Check that type names are PascalCase
		if !r.isPascalCase(def.Name) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Type name `%s` should be PascalCase.", def.Name),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
		//Type / Object
		if def.Kind == ast.Object && (strings.HasSuffix(def.Name, "Type") ||
			strings.HasSuffix(def.Name, "Object") ||
			strings.HasPrefix(def.Name, "Type") ||
			strings.HasPrefix(def.Name, "Object")) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Type name `%s` should be PascalCase and should not start/end with `Type` or `Object`", def.Name),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}

		if def.Kind == ast.Interface && (strings.HasSuffix(def.Name, "Interface") || strings.HasPrefix(def.Name, "Interface")) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Interface name `%s` should be PascalCase and should not start/end with `Interface`", def.Name),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	// Check field naming conventions
	for _, def := range schema.Types {
		if def.Kind == ast.Enum {
			// Check enum type name doesn't start/end with "Enum"
			line, column := 1, 1
			if def.Position != nil {
				line = def.Position.Line
				column = def.Position.Column
			}

			if strings.HasSuffix(strings.ToLower(def.Name), "enum") ||
				strings.HasPrefix(strings.ToLower(def.Name), "enum") {
				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Enum name `%s` should not start or end with `Enum`", def.Name),
					Location: types.Location{
						Line:   line,
						Column: column,
						File:   source.Name,
					},
					Rule: r.Name(),
				})
			}

			// Check enum values are UPPER_CASE
			for _, value := range def.EnumValues {
				valueLine, valueColumn := 1, 1
				if value.Position != nil {
					valueLine = value.Position.Line
					valueColumn = value.Position.Column
				}

				if !r.isUpperCase(value.Name) {
					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Enum value `%s.%s` should be UPPER_CASE", def.Name, value.Name),
						Location: types.Location{
							Line:   valueLine,
							Column: valueColumn,
							File:   source.Name,
						},
						Rule: r.Name(),
					})
				}
			}
		}
		if def.Kind != ast.Enum {
			for _, field := range def.Fields {
				// Skip built-in fields and introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}

				line, column := 1, 1
				if field.Position != nil {
					line = field.Position.Line
					column = field.Position.Column
				}

				// Check that field names are camelCase
				if !r.isCamelCase(field.Name) {
					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Field name `%s.%s` should be camelCase.", def.Name, field.Name),
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

// isPascalCase checks if a string follows PascalCase convention
func (r *NamingConvention) isPascalCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Must start with uppercase letter
	if s[0] < 'A' || s[0] > 'Z' {
		return false
	}

	// Check for valid PascalCase pattern
	pascalRegex := regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)
	return pascalRegex.MatchString(s)
}

// isCamelCase checks if a string follows camelCase convention
func (r *NamingConvention) isCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Must start with lowercase letter
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}

	// Check for valid camelCase pattern
	camelRegex := regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)
	return camelRegex.MatchString(s)
}

// isUpperCase checks if a string follows UPPER_CASE convention
func (r *NamingConvention) isUpperCase(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Check for valid UPPER_CASE pattern (letters, numbers, underscores)
	// Must contain at least one letter and all letters must be uppercase
	upperRegex := regexp.MustCompile(`^[A-Z0-9_]+$`)
	hasLetter := regexp.MustCompile(`[A-Z]`)

	return upperRegex.MatchString(s) && hasLetter.MatchString(s)
}
