package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// RelayEdgeTypes checks that Edge types follow the Relay specification
type RelayEdgeTypes struct{}

// NewRelayEdgeTypes creates a new instance of the RelayEdgeTypes rule
func NewRelayEdgeTypes() *RelayEdgeTypes {
	return &RelayEdgeTypes{}
}

// Name returns the rule name
func (r *RelayEdgeTypes) Name() string {
	return "relay-edge-types"
}

// Description returns what this rule checks
func (r *RelayEdgeTypes) Description() string {
	return "Ensure Edge types follow Relay specification - must be Object types with node and cursor fields, where node implements Node interface"
}

// Check validates that Edge types follow Relay specifications
func (r *RelayEdgeTypes) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// First, find all Connection types and collect Edge types referenced by them
	edgeTypes := make(map[string]bool)

	// Check all types in the schema
	for _, def := range schema.Types {
		// Skip built-in types and introspection types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		// Also check types that end with "Edge" by name convention
		if strings.HasSuffix(def.Name, "Edge") {
			edgeTypes[def.Name] = true
		}
	}

	// Validate each identified Edge type
	for edgeTypeName := range edgeTypes {
		if edgeTypeDef := schema.Types[edgeTypeName]; edgeTypeDef != nil {
			errors = append(errors, r.validateEdgeType(edgeTypeDef, schema, source)...)
		}
	}

	return errors
}

// validateEdgeType validates that an Edge type meets Relay specifications
func (r *RelayEdgeTypes) validateEdgeType(edgeType *ast.Definition, schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	line, column := 1, 1
	if edgeType.Position != nil {
		line = edgeType.Position.Line
		column = edgeType.Position.Column
	}

	// Rule 1: Edge type must be an Object type
	if edgeType.Kind != ast.Object {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Edge type `%s` must be an Object type, but is %s.",
				edgeType.Name, edgeType.Kind),
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

	// Rule 2: Edge type must contain a field node that returns either Scalar, Enum, Object, Interface, Union, or a non-null wrapper around one of those types. Notably, this field cannot return a list
	nodeField := r.findField(edgeType, "node")
	if nodeField == nil {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Edge type `%s` must contain a field `node` that returns either Scalar, Enum, Object, Interface, Union, or a non-null wrapper around one of those types.",
				edgeType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	} else {
		// Validate that node field doesn't return a list
		if r.isListType(nodeField.Type) {
			fieldLine, fieldColumn := 1, 1
			if nodeField.Position != nil {
				fieldLine = nodeField.Position.Line
				fieldColumn = nodeField.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Edge type `%s` field `node` cannot return a list type, but returns %s.",
					edgeType.Name, r.typeToString(nodeField.Type)),
				Location: types.Location{
					Line:   fieldLine,
					Column: fieldColumn,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}

		// Validate that node field type is valid (Scalar, Enum, Object, Interface, Union)
		nodeTypeName := r.getBaseTypeName(nodeField.Type)
		if nodeTypeName != "" {
			if nodeTypeDef := schema.Types[nodeTypeName]; nodeTypeDef != nil {
				if !r.isValidNodeFieldType(nodeTypeDef.Kind) {
					fieldLine, fieldColumn := 1, 1
					if nodeField.Position != nil {
						fieldLine = nodeField.Position.Line
						fieldColumn = nodeField.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Edge type `%s` field `node` must return Scalar, Enum, Object, Interface, or Union type, but returns %s.",
							edgeType.Name, nodeTypeDef.Kind),
						Location: types.Location{
							Line:   fieldLine,
							Column: fieldColumn,
							File:   source.Name,
						},
						Rule: r.Name(),
					})
				}
			}
		}
	}

	// Rule 3: Edge type must contain a field cursor that returns either String, Scalar, or a non-null wrapper around one of those types
	cursorField := r.findField(edgeType, "cursor")
	if cursorField == nil {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Edge type `%s` must contain a field `cursor` that returns either String, Scalar, or a non-null wrapper around one of those types.",
				edgeType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	} else {
		// Validate that cursor field returns appropriate type (String, Scalar, or non-null wrapper)
		if !r.isValidCursorFieldType(cursorField.Type, schema) {
			fieldLine, fieldColumn := 1, 1
			if cursorField.Position != nil {
				fieldLine = cursorField.Position.Line
				fieldColumn = cursorField.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Edge type `%s` field `cursor` must return String, Scalar, or a non-null wrapper around one of those types, but returns %s.",
					edgeType.Name, r.typeToString(cursorField.Type)),
				Location: types.Location{
					Line:   fieldLine,
					Column: fieldColumn,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	// Rule 4: Edge type name must end in "Edge"
	if !strings.HasSuffix(edgeType.Name, "Edge") {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Edge type `%s` name must end with 'Edge'.", edgeType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	// Rule 5: Edge type's field node must implement Node interface
	if nodeField != nil {
		nodeTypeName := r.getBaseTypeName(nodeField.Type)
		if nodeTypeName != "" {
			if nodeTypeDef := schema.Types[nodeTypeName]; nodeTypeDef != nil {
				if nodeTypeDef.Kind == ast.Object || nodeTypeDef.Kind == ast.Interface {
					if !r.implementsNodeInterface(nodeTypeDef, schema) {
						fieldLine, fieldColumn := 1, 1
						if nodeField.Position != nil {
							fieldLine = nodeField.Position.Line
							fieldColumn = nodeField.Position.Column
						}

						errors = append(errors, types.LintError{
							Message: fmt.Sprintf("Edge type `%s` field `node` type `%s` must implement Node interface.",
								edgeType.Name, nodeTypeName),
							Location: types.Location{
								Line:   fieldLine,
								Column: fieldColumn,
								File:   source.Name,
							},
							Rule: r.Name(),
						})
					}
				}
			}
		}
	}

	return errors
}

// findField finds a field by name in a type definition
func (r *RelayEdgeTypes) findField(typeDef *ast.Definition, fieldName string) *ast.FieldDefinition {
	for _, field := range typeDef.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

// isListType checks if a type is a list type (with or without NonNull wrapper)
func (r *RelayEdgeTypes) isListType(fieldType *ast.Type) bool {
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

// getBaseTypeName extracts the base type name from a GraphQL type (removes NonNull and List wrappers)
func (r *RelayEdgeTypes) getBaseTypeName(fieldType *ast.Type) string {
	if fieldType.NamedType != "" {
		return fieldType.NamedType
	}

	if fieldType.Elem != nil {
		return r.getBaseTypeName(fieldType.Elem)
	}

	return ""
}

// getListElementTypeName extracts the element type name from a list type
func (r *RelayEdgeTypes) getListElementTypeName(fieldType *ast.Type) string {
	// If it's a NonNull wrapper, check the inner type
	if fieldType.NonNull && fieldType.Elem != nil {
		return r.getListElementTypeName(fieldType.Elem)
	}

	// If it's a list, get the element type
	if fieldType.Elem != nil && fieldType.NamedType == "" {
		return r.getBaseTypeName(fieldType.Elem)
	}

	return ""
}

// isValidNodeFieldType checks if a type kind is valid for a node field
func (r *RelayEdgeTypes) isValidNodeFieldType(kind ast.DefinitionKind) bool {
	switch kind {
	case ast.Scalar, ast.Enum, ast.Object, ast.Interface, ast.Union:
		return true
	default:
		return false
	}
}

// isValidCursorFieldType checks if a type is a valid cursor field type (String, Scalar, or non-null wrapper)
func (r *RelayEdgeTypes) isValidCursorFieldType(fieldType *ast.Type, schema *ast.Schema) bool {
	// Check if it's directly a String type
	if fieldType.NamedType == "String" {
		return true
	}

	// Check if it's a Scalar type
	if fieldType.NamedType != "" {
		if scalarTypeDef := schema.Types[fieldType.NamedType]; scalarTypeDef != nil {
			if scalarTypeDef.Kind == ast.Scalar {
				return true
			}
		}
	}

	// Check if it's a NonNull wrapper around a valid cursor type
	if fieldType.NonNull && fieldType.Elem != nil {
		return r.isValidCursorFieldType(fieldType.Elem, schema)
	}

	return false
}

// implementsNodeInterface checks if a type implements the Node interface
func (r *RelayEdgeTypes) implementsNodeInterface(typeDef *ast.Definition, schema *ast.Schema) bool {
	// Check if Node interface exists in schema
	nodeInterface := schema.Types["Node"]
	if nodeInterface == nil || nodeInterface.Kind != ast.Interface {
		// If Node interface doesn't exist, we can't enforce this rule
		return true
	}

	// For Interface types, check if they implement Node
	if typeDef.Kind == ast.Interface {
		for _, impl := range typeDef.Interfaces {
			if impl == "Node" {
				return true
			}
		}
		return false
	}

	// For Object types, check if they implement Node interface
	if typeDef.Kind == ast.Object {
		for _, impl := range typeDef.Interfaces {
			if impl == "Node" {
				return true
			}
		}
		return false
	}

	return true // Other types don't need to implement Node
}

// typeToString converts a GraphQL type to its string representation
func (r *RelayEdgeTypes) typeToString(fieldType *ast.Type) string {
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
