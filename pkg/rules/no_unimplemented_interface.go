package rules

import (
	"fmt"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// NoUnimplementedInterface checks for interfaces that are not implemented by any type
type NoUnimplementedInterface struct{}

// NewNoUnimplementedInterface creates a new instance of NoUnimplementedInterface
func NewNoUnimplementedInterface() *NoUnimplementedInterface {
	return &NoUnimplementedInterface{}
}

// Name returns the name of this rule
func (r *NoUnimplementedInterface) Name() string {
	return "no-unimplemented-interface"
}

// Description returns the description of this rule
func (r *NoUnimplementedInterface) Description() string {
	return "Flags interfaces that are not implemented by any type in the schema"
}

// Check validates that all interfaces are implemented by at least one type
func (r *NoUnimplementedInterface) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find all interfaces in the schema
	interfaces := make(map[string]*ast.Definition)
	for typeName, typeDefinition := range schema.Types {
		// Skip built-in types
		if typeDefinition.BuiltIn {
			continue
		}

		if typeDefinition.Kind == ast.Interface {
			interfaces[typeName] = typeDefinition
		}
	}

	// Track which interfaces are implemented
	implementedInterfaces := make(map[string]bool)

	// Check all types to see which interfaces they implement
	for _, typeDefinition := range schema.Types {
		// Skip built-in types
		if typeDefinition.BuiltIn {
			continue
		}

		// Only object types and interface types can implement interfaces
		//if typeDefinition.Kind == ast.Object || typeDefinition.Kind == ast.Interface {
		for _, interfaceName := range typeDefinition.Interfaces {
			implementedInterfaces[interfaceName] = true
		}
		//}
	}

	// Flag interfaces that are not implemented
	for interfaceName, interfaceDefinition := range interfaces {
		if !implementedInterfaces[interfaceName] {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Interface '%s' is not implemented by any type", interfaceName),
				Location: types.Location{
					Line:   interfaceDefinition.Position.Line,
					Column: interfaceDefinition.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}
