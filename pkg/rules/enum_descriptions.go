package rules

import (
	"fmt"
	"strings"

	"github.com/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// EnumDescriptions checks that enum members have descriptions except for UNKNOWN
type EnumDescriptions struct{}

// NewEnumDescriptions creates a new instance of the EnumDescriptions rule
func NewEnumDescriptions() *EnumDescriptions {
	return &EnumDescriptions{}
}

// Name returns the rule name
func (r *EnumDescriptions) Name() string {
	return "enum-descriptions"
}

// Description returns what this rule checks
func (r *EnumDescriptions) Description() string {
	return "All enum values must have descriptions except for UNKNOWN case"
}

// Check validates that enum values have descriptions
func (r *EnumDescriptions) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all enum types
	for _, def := range schema.Types {
		if def.Kind != ast.Enum {
			continue
		}

		// Skip introspection enums
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		// Check each enum value
		for _, enumValue := range def.EnumValues {
			// Skip the UNKNOWN case - it doesn't need a description
			if enumValue.Name == "UNKNOWN" {
				continue
			}

			// Check if the enum value has a description
			if enumValue.Description == "" {
				line, column := 1, 1
				if enumValue.Position != nil {
					line = enumValue.Position.Line
					column = enumValue.Position.Column
				}

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Enum value `%s.%s` is missing a description. All enum values except UNKNOWN should have descriptions.", def.Name, enumValue.Name),
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
