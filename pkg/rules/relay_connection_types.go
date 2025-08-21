package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
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
		lowerCaseDefName := strings.ToLower(def.Name)
		// Check if this is a Connection type (ends with "Connection")
		if strings.HasSuffix(lowerCaseDefName, "connection") {
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
		return errors
	}

	// Rule 2: Connection type must contain an "edges" field that returns a list type
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
		// TODO: Do we need to add a check that edges object name is <prefix>Edge?
		// Validate that edges field returns a single-level list type
		fieldLine, fieldColumn := 1, 1
		if edgesField.Position != nil {
			fieldLine = edgesField.Position.Line
			fieldColumn = edgesField.Position.Column
		}

		if !isListType(edgesField.Type) {
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
		} else if isNestedListType(edgesField.Type) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Connection type `%s` field `edges` must return a single-level list type, but returns a nested list %s.",
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
	pageInfoField := r.findField(connectionType, "pageInfo")
	// TODO: Do we need to add a check that pageInfo object name is PageInfo?
	// pageInfo: randomName! is allowed?
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

// typeToString converts a GraphQL type to its string representation
func (r *RelayConnectionTypes) typeToString(fieldType *ast.Type) string {
	// Handle the base case - named type
	if fieldType.NamedType != "" {
		if fieldType.NonNull {
			return fieldType.NamedType + "!"
		}
		return fieldType.NamedType
	}

	// Handle list types
	if fieldType.Elem != nil {
		innerType := r.typeToString(fieldType.Elem)
		listType := "[" + innerType + "]"
		if fieldType.NonNull {
			return listType + "!"
		}
		return listType
	}

	return "Unknown"
}
