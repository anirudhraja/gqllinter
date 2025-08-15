package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// NoUnusedFields checks for fields that are never used/referenced
type NoUnusedFields struct{}

// NewNoUnusedFields creates a new instance of the NoUnusedFields rule
func NewNoUnusedFields() *NoUnusedFields {
	return &NoUnusedFields{}
}

// Name returns the rule name
func (r *NoUnusedFields) Name() string {
	return "no-unused-fields"
}

// Description returns what this rule checks
func (r *NoUnusedFields) Description() string {
	return "Detect unused fields in schema - following Guild best practices for clean schemas"
}

// Check validates that fields are actually used in the schema
func (r *NoUnusedFields) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Track all field references across the schema
	usedFields := make(map[string]map[string]bool) // typeName -> fieldName -> used

	// Initialize tracking for all types and their fields
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			usedFields[def.Name] = make(map[string]bool)
			for _, field := range def.Fields {
				usedFields[def.Name][field.Name] = false
			}
		}
	}

	// Mark fields as used based on various contexts
	r.markUsedFields(schema, usedFields)

	// Report unused fields
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			// Skip root types (Query, Mutation, Subscription) - their fields are entry points
			if def == schema.Query || def == schema.Mutation || def == schema.Subscription {
				continue
			}

			for _, field := range def.Fields {
				if !usedFields[def.Name][field.Name] {
					line, column := 1, 1
					if field.Position != nil {
						line = field.Position.Line
						column = field.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Field `%s.%s` is never used and can be removed.", def.Name, field.Name),
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

// markUsedFields traverses the schema to find field references
func (r *NoUnusedFields) markUsedFields(schema *ast.Schema, usedFields map[string]map[string]bool) {
	// Mark fields used in type references
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			for _, field := range def.Fields {
				// Mark fields referenced by field types
				r.markFieldsInType(field.Type, usedFields)

				// Mark fields referenced in arguments
				for _, arg := range field.Arguments {
					r.markFieldsInType(arg.Type, usedFields)
				}
			}
		}

		// Mark fields in input types
		if def.Kind == ast.InputObject {
			for _, field := range def.Fields {
				r.markFieldsInType(field.Type, usedFields)
			}
		}

		// Mark fields in union types
		if def.Kind == ast.Union {
			for _, memberType := range def.Types {
				if typeFields, exists := usedFields[memberType]; exists {
					// Mark all fields of union member types as potentially used
					for fieldName := range typeFields {
						usedFields[memberType][fieldName] = true
					}
				}
			}
		}

		// Mark fields in interface implementations
		if def.Kind == ast.Object {
			for _, interfaceName := range def.Interfaces {
				if interfaceType := schema.Types[interfaceName]; interfaceType != nil {
					for _, interfaceField := range interfaceType.Fields {
						// Mark corresponding fields in implementing types
						if _, exists := usedFields[def.Name][interfaceField.Name]; exists {
							usedFields[def.Name][interfaceField.Name] = true
						}
					}
				}
			}
		}
	}
}

// markFieldsInType marks fields that are referenced by a type
func (r *NoUnusedFields) markFieldsInType(fieldType *ast.Type, usedFields map[string]map[string]bool) {
	// Get the base type name (unwrap lists and non-nulls)
	baseType := fieldType
	for baseType.Elem != nil {
		baseType = baseType.Elem
	}

	typeName := baseType.Name()

	// If this type has fields we're tracking, mark them as potentially used
	if typeFields, exists := usedFields[typeName]; exists {
		for fieldName := range typeFields {
			usedFields[typeName][fieldName] = true
		}
	}
}
