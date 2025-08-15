package rules

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// CapitalizedDescriptions checks that all descriptions start with capital letters
type CapitalizedDescriptions struct{}

// NewCapitalizedDescriptions creates a new instance of the CapitalizedDescriptions rule
func NewCapitalizedDescriptions() *CapitalizedDescriptions {
	return &CapitalizedDescriptions{}
}

// Name returns the rule name
func (r *CapitalizedDescriptions) Name() string {
	return "capitalized-descriptions"
}

// Description returns what this rule checks
func (r *CapitalizedDescriptions) Description() string {
	return "All descriptions must start with a capital letter for consistency"
}

// Check validates that all descriptions are properly capitalized
func (r *CapitalizedDescriptions) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check type descriptions
	for _, def := range schema.Types {
		// Skip built-in types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		if def.Description != "" && !r.isCapitalized(def.Description) {
			line, column := 1, 1
			if def.Position != nil {
				line = def.Position.Line
				column = def.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Description for type `%s` should start with a capital letter.", def.Name),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}

		// Check field descriptions
		for _, field := range def.Fields {
			// Skip built-in fields and introspection fields
			if strings.HasPrefix(field.Name, "__") {
				continue
			}

			if field.Description != "" && !r.isCapitalized(field.Description) {
				line, column := 1, 1
				if field.Position != nil {
					line = field.Position.Line
					column = field.Position.Column
				}

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Description for field `%s.%s` should start with a capital letter.", def.Name, field.Name),
					Location: types.Location{
						Line:   line,
						Column: column,
						File:   source.Name,
					},
					Rule: r.Name(),
				})
			}

			// Check field argument descriptions
			for _, arg := range field.Arguments {
				if arg.Description != "" && !r.isCapitalized(arg.Description) {
					line, column := 1, 1
					if arg.Position != nil {
						line = arg.Position.Line
						column = arg.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Description for argument `%s.%s(%s:)` should start with a capital letter.", def.Name, field.Name, arg.Name),
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

		// Check enum value descriptions
		if def.Kind == ast.Enum {
			for _, enumValue := range def.EnumValues {
				if enumValue.Description != "" && !r.isCapitalized(enumValue.Description) {
					line, column := 1, 1
					if enumValue.Position != nil {
						line = enumValue.Position.Line
						column = enumValue.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Description for enum value `%s.%s` should start with a capital letter.", def.Name, enumValue.Name),
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

	// Check directive descriptions
	for _, directive := range schema.Directives {
		if directive.Description != "" && !r.isCapitalized(directive.Description) {
			line, column := 1, 1
			if directive.Position != nil {
				line = directive.Position.Line
				column = directive.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Description for directive `@%s` should start with a capital letter.", directive.Name),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}

		// Check directive argument descriptions
		for _, arg := range directive.Arguments {
			if arg.Description != "" && !r.isCapitalized(arg.Description) {
				line, column := 1, 1
				if arg.Position != nil {
					line = arg.Position.Line
					column = arg.Position.Column
				}

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Description for directive argument `@%s(%s:)` should start with a capital letter.", directive.Name, arg.Name),
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

// isCapitalized checks if a description starts with a capital letter
func (r *CapitalizedDescriptions) isCapitalized(description string) bool {
	// Trim whitespace and check if the first character is uppercase
	trimmed := strings.TrimSpace(description)
	if len(trimmed) == 0 {
		return true // Empty descriptions are fine
	}

	firstChar := rune(trimmed[0])
	return unicode.IsUpper(firstChar)
}
