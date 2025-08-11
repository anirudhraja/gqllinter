package rules

import (
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// NoInvalidEnum checks that no enum value is named "INVALID"
type NoInvalidEnum struct{}

// NewNoInvalidEnum creates a new instance of the NoInvalidEnum rule
func NewNoInvalidEnum() *NoInvalidEnum {
	return &NoInvalidEnum{}
}

// Name returns the rule name
func (r *NoInvalidEnum) Name() string {
	return "no-invalid-enum"
}

// Description returns what this rule checks
func (r *NoInvalidEnum) Description() string {
	return "Ensures that no enum value is named 'INVALID' to avoid conflicts with proto translation zero values"
}

// Check validates that no enum values are named "INVALID"
func (r *NoInvalidEnum) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all enum definitions in the schema
	for _, def := range schema.Types {
		if def.Kind == ast.Enum {
			errors = append(errors, r.checkEnumValues(def, source)...)
		}
	}

	return errors
}

// checkEnumValues checks if any enum value is named "INVALID"
func (r *NoInvalidEnum) checkEnumValues(enumDef *ast.Definition, source *ast.Source) []types.LintError {
	var errors []types.LintError

	for _, enumValue := range enumDef.EnumValues {
		if strings.ToUpper(enumValue.Name) == "INVALID" {
			errors = append(errors, types.LintError{
				Message: "Enum value 'INVALID' is not allowed as it conflicts with proto translation zero values. Use a different name for enum '" + enumDef.Name + "'",
				Location: types.Location{
					Line:   enumValue.Position.Line,
					Column: enumValue.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}
