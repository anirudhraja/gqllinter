package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// RelayConnectionTypes checks that Connection types follow the Relay specification
type RelayConnectionTypes struct{}

// NewRelayConnectionTypes creates a new instance of the RelayConnectionTypes rule
func NewRelayConnectionTypes() *RelayConnectionTypes {
	return &RelayConnectionTypes{}
}

// Name returns the rule name
func (r *RelayConnectionTypes) Name() string {
	return "relay-connection-types"
}

// Description returns what this rule checks
func (r *RelayConnectionTypes) Description() string {
	return "Ensure Connection types follow Relay specification - must be Object types with edges and pageInfo fields"
}

// Check validates that Connection types follow Relay specifications
// TODO (bishnu.agrawal) - add test cases for this rule
func (r *RelayConnectionTypes) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all types in the schema
	for _, def := range schema.Types {
		// Skip built-in types and introspection types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		// Check if this is a Connection type (ends with "Connection")
		if strings.HasSuffix(def.Name, "Connection") {
			errors = append(errors, r.validateConnectionType(def, source)...)
		}
	}

	return errors
}

// validateConnectionType validates that a Connection type meets Relay specifications
func (r *RelayConnectionTypes) validateConnectionType(connectionType *ast.Definition, source *ast.Source) []types.LintError {
	var errors []types.LintError

	line, column := 1, 1
	if connectionType.Position != nil {
		line = connectionType.Position.Line
		column = connectionType.Position.Column
	}

	// Rule 1: Connection type must be an Object type
	if connectionType.Kind != ast.Object {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Connection type `%s` must be an Object type, but is %s.",
				connectionType.Name, connectionType.Kind),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
		// If it's not an Object type, we can't check fields
		return errors
	}

	// Rule 2: Connection type must contain an "edges" field that returns a list type
	// TODO (bishnu.agrawal) - revisit the rule
	edgesField := r.findField(connectionType, "edges")
	if edgesField == nil {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Connection type `%s` must contain a field `edges` that returns a list type.",
				connectionType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	} else {
		// Validate that edges field returns a list type
		if !r.isListType(edgesField.Type) {
			fieldLine, fieldColumn := 1, 1
			if edgesField.Position != nil {
				fieldLine = edgesField.Position.Line
				fieldColumn = edgesField.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Connection type `%s` field `edges` must return a list type, but returns %s.",
					connectionType.Name, r.typeToString(edgesField.Type)),
				Location: types.Location{
					Line:   fieldLine,
					Column: fieldColumn,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}

		// edgesField must satisafy all the rules of RelayEdgeTypes
		//edgeErrors := NewRelayEdgeTypes().Check(schema, source)
		//errors = append(errors, edgeErrors...)
	}

	// Rule 3: Connection type must contain a "pageInfo" field that returns non-null PageInfo
	// TODO (bishnu.agrawal) - revisit the rule
	pageInfoField := r.findField(connectionType, "pageInfo")
	// || pageInfoField.Type.NamedType != ast.Object - TODO
	if pageInfoField == nil || !pageInfoField.Type.NonNull {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Connection type `%s` must contain a field `pageInfo` that returns a non-null PageInfo Object type.",
				connectionType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	} else {
		// Additionally validate that the PageInfo type itself follows Relay specification
		// TODO
		//pageInfoErrors := NewRelayPageInfo().ValidatePageInfoType(pageInfoField, source)
		//errors = append(errors, pageInfoErrors...)
	}

	return errors
}

// findField finds a field by name in a type definition
func (r *RelayConnectionTypes) findField(typeDef *ast.Definition, fieldName string) *ast.FieldDefinition {
	for _, field := range typeDef.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

// isListType checks if a type is a list type (with or without NonNull wrapper)
func (r *RelayConnectionTypes) isListType(fieldType *ast.Type) bool {
	// Check if it's directly a list
	if fieldType.Elem != nil && fieldType.NamedType == "" {
		return true
	}

	// Check if it's a NonNull wrapper around a list
	if fieldType.NonNull && fieldType.Elem != nil {
		return r.isListType(fieldType.Elem)
	}

	return false
}

// typeToString converts a GraphQL type to its string representation
func (r *RelayConnectionTypes) typeToString(fieldType *ast.Type) string {
	if fieldType.NonNull {
		if fieldType.NamedType != "" {
			return fieldType.NamedType + "!"
		}
		if fieldType.Elem != nil {
			return r.typeToString(fieldType.Elem) + "!"
		}
	}

	if fieldType.NamedType != "" {
		return fieldType.NamedType
	}

	if fieldType.Elem != nil {
		return "[" + r.typeToString(fieldType.Elem) + "]"
	}

	return "Unknown"
}
