package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// EnumUnknownCase checks that all enums used in output types have an UNKNOWN case
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
	return "All enums used in output types (not inputs) must have an UNKNOWN case for future compatibility"
}

// Check validates that output enums have UNKNOWN cases
func (r *EnumUnknownCase) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find all enums used in output types
	outputEnums := r.findOutputEnums(schema)

	// Check each enum used in output types
	for enumName := range outputEnums {
		enumDef := schema.Types[enumName]
		if enumDef == nil || enumDef.Kind != ast.Enum {
			continue
		}

		// Skip introspection enums
		if strings.HasPrefix(enumName, "__") {
			continue
		}

		// Check if this enum has an UNKNOWN case
		hasUnknown := false
		for _, enumValue := range enumDef.EnumValues {
			if enumValue.Name == "UNKNOWN" {
				hasUnknown = true
				break
			}
		}

		if !hasUnknown {
			line, column := 1, 1
			if enumDef.Position != nil {
				line = enumDef.Position.Line
				column = enumDef.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Enum `%s` is used in output types but lacks an UNKNOWN case. Add `UNKNOWN` for future compatibility when new enum values are introduced.", enumName),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}

// findOutputEnums identifies all enums that are used in output types (not input types)
func (r *EnumUnknownCase) findOutputEnums(schema *ast.Schema) map[string]bool {
	outputEnums := make(map[string]bool)

	// Check all non-input types
	for _, def := range schema.Types {
		// Skip input types - they don't need UNKNOWN cases
		if def.Kind == ast.InputObject {
			continue
		}

		// Skip introspection types
		if strings.HasPrefix(def.Name, "__") {
			continue
		}

		// Check object, interface, and union types
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.Union {
			for _, field := range def.Fields {
				// Skip built-in fields and introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}

				enumType := r.getBaseTypeName(field.Type)
				if r.isEnum(schema, enumType) {
					outputEnums[enumType] = true
				}

				// Check field arguments (these are technically input, but field args in output types should still have UNKNOWN)
				for _, arg := range field.Arguments {
					argEnumType := r.getBaseTypeName(arg.Type)
					if r.isEnum(schema, argEnumType) {
						outputEnums[argEnumType] = true
					}
				}
			}
		}

		// Check union member types
		if def.Kind == ast.Union {
			for _, memberType := range def.Types {
				if r.isEnum(schema, memberType) {
					outputEnums[memberType] = true
				}
			}
		}
	}

	// Also check root types
	rootTypes := []*ast.Definition{schema.Query, schema.Mutation, schema.Subscription}
	for _, rootType := range rootTypes {
		if rootType == nil {
			continue
		}

		for _, field := range rootType.Fields {
			// Skip built-in fields and introspection fields
			if strings.HasPrefix(field.Name, "__") {
				continue
			}

			enumType := r.getBaseTypeName(field.Type)
			if r.isEnum(schema, enumType) {
				outputEnums[enumType] = true
			}
		}
	}

	return outputEnums
}

// isEnum checks if a type name refers to an enum type
func (r *EnumUnknownCase) isEnum(schema *ast.Schema, typeName string) bool {
	if typeDef := schema.Types[typeName]; typeDef != nil {
		return typeDef.Kind == ast.Enum
	}
	return false
}

// getBaseTypeName extracts the base type name from a type reference
func (r *EnumUnknownCase) getBaseTypeName(fieldType *ast.Type) string {
	// Unwrap lists and non-nulls to get the base type
	baseType := fieldType
	for baseType.Elem != nil {
		baseType = baseType.Elem
	}
	return baseType.Name()
}
