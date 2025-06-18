package rules

import (
	"fmt"
	"strings"

	"github.com/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// InputEnumSuffix checks that input enums are distinct from output enums and are suffixed with "Input"
type InputEnumSuffix struct{}

// NewInputEnumSuffix creates a new instance of the InputEnumSuffix rule
func NewInputEnumSuffix() *InputEnumSuffix {
	return &InputEnumSuffix{}
}

// Name returns the rule name
func (r *InputEnumSuffix) Name() string {
	return "input-enum-suffix"
}

// Description returns what this rule checks
func (r *InputEnumSuffix) Description() string {
	return "Input enums must be distinct from output enums and suffixed with 'Input' for clarity"
}

// Check validates that input enums follow the naming convention
func (r *InputEnumSuffix) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find all enums used in input contexts
	inputEnums := r.findInputEnums(schema)
	outputEnums := r.findOutputEnums(schema)

	// Check each enum used in input contexts
	for enumName := range inputEnums {
		enumDef := schema.Types[enumName]
		if enumDef == nil || enumDef.Kind != ast.Enum {
			continue
		}

		// Skip introspection enums
		if strings.HasPrefix(enumName, "__") {
			continue
		}

		// If this enum is used in both input and output contexts, it should be split
		if outputEnums[enumName] {
			line, column := 1, 1
			if enumDef.Position != nil {
				line = enumDef.Position.Line
				column = enumDef.Position.Column
			}

			inputEnumName := enumName
			if !strings.HasSuffix(enumName, "Input") {
				inputEnumName = enumName + "Input"
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Enum `%s` is used in both input and output contexts. Consider creating separate input enum `%s` for input usage.", enumName, inputEnumName),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		} else {
			// This enum is only used in input contexts - it should end with "Input"
			if !strings.HasSuffix(enumName, "Input") {
				line, column := 1, 1
				if enumDef.Position != nil {
					line = enumDef.Position.Line
					column = enumDef.Position.Column
				}

				suggestedName := enumName + "Input"

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Input enum `%s` should be suffixed with 'Input'. Consider renaming to `%s`.", enumName, suggestedName),
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

// findInputEnums identifies all enums that are used in input contexts
func (r *InputEnumSuffix) findInputEnums(schema *ast.Schema) map[string]bool {
	inputEnums := make(map[string]bool)

	// Check input object types
	for _, def := range schema.Types {
		if def.Kind == ast.InputObject {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			for _, field := range def.Fields {
				enumType := r.getBaseTypeName(field.Type)
				if r.isEnum(schema, enumType) {
					inputEnums[enumType] = true
				}
			}
		}
	}

	// Check field arguments (these are input contexts)
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			for _, field := range def.Fields {
				// Skip introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}

				for _, arg := range field.Arguments {
					enumType := r.getBaseTypeName(arg.Type)
					if r.isEnum(schema, enumType) {
						inputEnums[enumType] = true
					}
				}
			}
		}
	}

	// Check directive arguments
	for _, directive := range schema.Directives {
		for _, arg := range directive.Arguments {
			enumType := r.getBaseTypeName(arg.Type)
			if r.isEnum(schema, enumType) {
				inputEnums[enumType] = true
			}
		}
	}

	return inputEnums
}

// findOutputEnums identifies all enums that are used in output contexts
func (r *InputEnumSuffix) findOutputEnums(schema *ast.Schema) map[string]bool {
	outputEnums := make(map[string]bool)

	// Check all non-input types
	for _, def := range schema.Types {
		// Skip input types
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
				// Skip introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}

				enumType := r.getBaseTypeName(field.Type)
				if r.isEnum(schema, enumType) {
					outputEnums[enumType] = true
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

	return outputEnums
}

// isEnum checks if a type name refers to an enum type
func (r *InputEnumSuffix) isEnum(schema *ast.Schema, typeName string) bool {
	if typeDef := schema.Types[typeName]; typeDef != nil {
		return typeDef.Kind == ast.Enum
	}
	return false
}

// getBaseTypeName extracts the base type name from a type reference
func (r *InputEnumSuffix) getBaseTypeName(fieldType *ast.Type) string {
	// Unwrap lists and non-nulls to get the base type
	baseType := fieldType
	for baseType.Elem != nil {
		baseType = baseType.Elem
	}
	return baseType.Name()
}
