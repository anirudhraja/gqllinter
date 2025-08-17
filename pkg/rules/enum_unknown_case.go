package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// EnumUnknownCase checks that enums do not contain UNKNOWN values
type EnumUnknownCase struct{}

// NewEnumUnknownCase creates a new instance of the EnumUnknownCase rule
func NewEnumUnknownCase() *EnumUnknownCase {
	return &EnumUnknownCase{}
}

// Name returns the rule name
func (r *EnumUnknownCase) Name() string {
	return "enum-unknown-case"
}

// Description returns what this rule checks
func (r *EnumUnknownCase) Description() string {
	return "Enums should not contain UNKNOWN values as they can lead to unclear business logic"
}

// Check validates that enums do not have UNKNOWN values
func (r *EnumUnknownCase) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all enums in the schema
	for enumName, enumDef := range schema.Types {
		if enumDef == nil || enumDef.Kind != ast.Enum {
			continue
		}

		// Skip introspection enums
		if strings.HasPrefix(enumName, "__") {
			continue
		}

		// Check if this enum has an UNKNOWN case and flag it
		for _, enumValue := range enumDef.EnumValues {
			if enumValue.Name == "UNKNOWN" {
				line, column := 1, 1
				if enumValue.Position != nil {
					line = enumValue.Position.Line
					column = enumValue.Position.Column
				} else if enumDef.Position != nil {
					line = enumDef.Position.Line
					column = enumDef.Position.Column
				}

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Enum `%s` contains an UNKNOWN value. UNKNOWN as a enum value is not allowed.", enumName),
					Location: types.Location{
						Line:   line,
						Column: column,
						File:   source.Name,
					},
					Rule: r.Name(),
				})
				break
			}
		}
	}

	return errors
}
