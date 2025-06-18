package rules

import (
	"fmt"
	"strings"

	"github.com/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// NoUnusedTypes checks that all declared types are actually used
type NoUnusedTypes struct{}

// NewNoUnusedTypes creates a new instance of the NoUnusedTypes rule
func NewNoUnusedTypes() *NoUnusedTypes {
	return &NoUnusedTypes{}
}

// Name returns the rule name
func (r *NoUnusedTypes) Name() string {
	return "no-unused-types"
}

// Description returns what this rule checks
func (r *NoUnusedTypes) Description() string {
	return "All declared types must be used somewhere in the schema - custom rule to support Federation"
}

// Check validates that all types are referenced somewhere
func (r *NoUnusedTypes) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Track which types are used
	usedTypes := make(map[string]bool)

	// Mark root types as used
	if schema.Query != nil {
		usedTypes[schema.Query.Name] = true
	}
	if schema.Mutation != nil {
		usedTypes[schema.Mutation.Name] = true
	}
	if schema.Subscription != nil {
		usedTypes[schema.Subscription.Name] = true
	}

	// Mark built-in types as used
	builtInTypes := []string{
		"String", "Int", "Float", "Boolean", "ID", "__Schema", "__Type", "__Field",
		"__InputValue", "__EnumValue", "__Directive", "__DirectiveLocation", "__TypeKind",
	}
	for _, typeName := range builtInTypes {
		usedTypes[typeName] = true
	}

	// Find all type references throughout the schema
	r.markTypeUsages(schema, usedTypes)

	// Check which types are unused
	for _, def := range schema.Types {
		// Skip built-in types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		if !usedTypes[def.Name] {
			line, column := 1, 1
			if def.Position != nil {
				line = def.Position.Line
				column = def.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Type `%s` is declared but never used. Consider removing it or using it in the schema.", def.Name),
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

// markTypeUsages recursively finds and marks all type usages
func (r *NoUnusedTypes) markTypeUsages(schema *ast.Schema, usedTypes map[string]bool) {
	// Keep iterating until no new types are found
	foundNew := true
	for foundNew {
		foundNew = false

		for _, def := range schema.Types {
			if !usedTypes[def.Name] {
				continue // Skip types we haven't marked as used yet
			}

			// Check fields in this type
			for _, field := range def.Fields {
				typeName := r.getBaseTypeName(field.Type)
				if !usedTypes[typeName] {
					usedTypes[typeName] = true
					foundNew = true
				}

				// Check field arguments
				for _, arg := range field.Arguments {
					argTypeName := r.getBaseTypeName(arg.Type)
					if !usedTypes[argTypeName] {
						usedTypes[argTypeName] = true
						foundNew = true
					}
				}
			}

			// Check interfaces implemented by this type
			for _, iface := range def.Interfaces {
				if !usedTypes[iface] {
					usedTypes[iface] = true
					foundNew = true
				}
			}

			// Check union member types
			if def.Kind == ast.Union {
				for _, memberType := range def.Types {
					if !usedTypes[memberType] {
						usedTypes[memberType] = true
						foundNew = true
					}
				}
			}
		}

		// Check directive definitions
		for _, directive := range schema.Directives {
			if !usedTypes[directive.Name] {
				continue
			}

			for _, arg := range directive.Arguments {
				argTypeName := r.getBaseTypeName(arg.Type)
				if !usedTypes[argTypeName] {
					usedTypes[argTypeName] = true
					foundNew = true
				}
			}
		}
	}
}

// getBaseTypeName extracts the base type name from a type reference
func (r *NoUnusedTypes) getBaseTypeName(fieldType *ast.Type) string {
	// Unwrap lists and non-nulls to get the base type
	baseType := fieldType
	for baseType.Elem != nil {
		baseType = baseType.Elem
	}
	return baseType.Name()
}
