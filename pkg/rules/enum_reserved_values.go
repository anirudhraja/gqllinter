package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// EnumReservedValues checks that enum values don't use reserved names
type EnumReservedValues struct{}

// NewEnumReservedValues creates a new instance of the EnumReservedValues rule
func NewEnumReservedValues() *EnumReservedValues {
	return &EnumReservedValues{}
}

// Name returns the rule name
func (r *EnumReservedValues) Name() string {
	return "enum-reserved-values"
}

// Description returns what this rule checks
func (r *EnumReservedValues) Description() string {
	return "Prevent use of reserved enum values for extensibility and future compatibility"
}

// Check validates that enum values don't use reserved names
func (r *EnumReservedValues) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Reserved enum values for future compatibility
	reservedValues := []string{"INVALID"}

	// Check enum types
	for _, def := range schema.Types {
		if def.Kind == ast.Enum {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			for _, enumValue := range def.EnumValues {
				if r.isReservedValue(enumValue.Name, reservedValues) {
					line, column := 1, 1
					if enumValue.Position != nil {
						line = enumValue.Position.Line
						column = enumValue.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Enum value `%s.%s` uses a reserved name.", def.Name, enumValue.Name),
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

// isReservedValue checks if an enum value is in the reserved list
func (r *EnumReservedValues) isReservedValue(value string, reserved []string) bool {
	valueUpper := strings.ToUpper(value)

	for _, reservedValue := range reserved {
		if valueUpper == reservedValue {
			return true
		}
	}

	return false
}
