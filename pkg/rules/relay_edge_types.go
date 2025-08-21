package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
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

	// Collect Edge types from two sources:
	// 1. Types referenced by Connection types' edges fields
	// 2. Types that end with "Edge" by name convention
	edgeTypes := make(map[string]bool)
	edgeTypesFromConnections := make(map[string]bool) // Track which were found from connections

	// Check all types in the schema
	for _, def := range schema.Types {
		// Skip built-in types and introspection types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}
		lowerCaseDefName := strings.ToLower(def.Name)
		// Check if this is a Connection type and extract referenced Edge types
		if strings.HasSuffix(lowerCaseDefName, "connection") && def.Kind == ast.Object {
			edgesField := r.findField(def, "edges")
			if edgesField != nil {
				edgeTypeName := r.getEdgeTypeFromEdgesField(edgesField.Type)
				if edgeTypeName != "" {
					edgeTypes[edgeTypeName] = true
					edgeTypesFromConnections[edgeTypeName] = true
				}
			}
		}

		// Also check types that end with "Edge" by name convention
		//TODO: Do we need to check these?
		//if strings.HasSuffix(def.Name, "Edge") {
		//	edgeTypes[def.Name] = true
		//}
	}

	// Validate each identified Edge type
	for edgeTypeName := range edgeTypes {
		if edgeTypeDef := schema.Types[edgeTypeName]; edgeTypeDef != nil {
			isFromConnection := edgeTypesFromConnections[edgeTypeName]
			errors = append(errors, r.validateEdgeType(edgeTypeDef, schema, source, isFromConnection)...)
		}
	}

	return errors
}

// validateEdgeType validates that an Edge type meets Relay specifications
func (r *RelayEdgeTypes) validateEdgeType(edgeType *ast.Definition, schema *ast.Schema, source *ast.Source, isFromConnection bool) []types.LintError {
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
					edgeType.Name, nodeField.Type.String()),
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
		// Validate that cursor field returns appropriate type (String, or non-null String wrapper)
		if cursorField.Type.NamedType != "String" {
			fieldLine, fieldColumn := 1, 1
			if cursorField.Position != nil {
				fieldLine = cursorField.Position.Line
				fieldColumn = cursorField.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Edge type `%s` field `cursor` must return String, or a non-null wrapper around a String, but returns %s.",
					edgeType.Name, cursorField.Type.String()),
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
	// Only enforce this rule for types that were detected by name convention, not from Connection edges
	if !isFromConnection && !strings.HasSuffix(edgeType.Name, "Edge") {
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
				} else if nodeTypeDef.Kind == ast.Union {
					// For union types, check if all union members implement Node interface
					unionErrors := r.validateUnionMembersImplementNode(nodeTypeDef, edgeType.Name, nodeField, schema, source)
					errors = append(errors, unionErrors...)
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

// isValidNodeFieldType checks if a type kind is valid for a node field
func (r *RelayEdgeTypes) isValidNodeFieldType(kind ast.DefinitionKind) bool {
	switch kind {
	case ast.Scalar, ast.Enum, ast.Object, ast.Interface, ast.Union:
		return true
	default:
		return false
	}
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

// validateUnionMembersImplementNode checks if all union members implement Node interface
func (r *RelayEdgeTypes) validateUnionMembersImplementNode(unionDef *ast.Definition, edgeTypeName string, nodeField *ast.FieldDefinition, schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check if Node interface exists in schema
	nodeInterface := schema.Types["Node"]
	if nodeInterface == nil || nodeInterface.Kind != ast.Interface {
		// If Node interface doesn't exist, we can't enforce this rule
		return errors
	}

	fieldLine, fieldColumn := 1, 1
	if nodeField.Position != nil {
		fieldLine = nodeField.Position.Line
		fieldColumn = nodeField.Position.Column
	}

	// Check each union member
	for _, memberName := range unionDef.Types {
		if memberDef := schema.Types[memberName]; memberDef != nil {
			if memberDef.Kind == ast.Object || memberDef.Kind == ast.Interface {
				if !r.implementsNodeInterface(memberDef, schema) {
					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Edge type `%s` field `node` union type `%s` member `%s` must implement Node interface.",
							edgeTypeName, unionDef.Name, memberName),
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

	return errors
}

// getEdgeTypeFromEdgesField extracts the Edge type name from an edges field type
// Handles cases like [UserEdge], [UserEdge!], [UserEdge]!, [UserEdge!]!
func (r *RelayEdgeTypes) getEdgeTypeFromEdgesField(fieldType *ast.Type) string {
	// If it's a NonNull wrapper, look at the inner type
	if fieldType.NonNull && fieldType.Elem != nil {
		return r.getEdgeTypeFromEdgesField(fieldType.Elem)
	}

	// If it's a list type, get the element type
	if fieldType.Elem != nil && fieldType.NamedType == "" {
		// This is a list, get the element type
		elementType := fieldType.Elem

		// Handle NonNull element (e.g., [UserEdge!])
		if elementType.NonNull && elementType.Elem != nil {
			return r.getEdgeTypeFromEdgesField(elementType.Elem)
		}

		// Handle regular element (e.g., [UserEdge])
		if elementType.NamedType != "" {
			return elementType.NamedType
		}
	}

	// Direct named type (shouldn't happen for edges field but handle it)
	if fieldType.NamedType != "" {
		return fieldType.NamedType
	}

	return ""
}
