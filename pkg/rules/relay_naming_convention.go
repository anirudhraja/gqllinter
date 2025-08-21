package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// RelayNamingConvention checks that Connection and Edge types follow Relay naming conventions
type RelayNamingConvention struct{}

// NewRelayNamingConvention creates a new instance of the RelayNamingConvention rule
func NewRelayNamingConvention() *RelayNamingConvention {
	return &RelayNamingConvention{}
}

// Name returns the rule name
func (r *RelayNamingConvention) Name() string {
	return "relay-naming-convention"
}

// Description returns what this rule checks
func (r *RelayNamingConvention) Description() string {
	return "Ensure Connection and Edge types follow Relay naming conventions: Connection must be named [Entity]Connection with edges field of type [Entity]Edge, Edge must be named [Entity]Edge"
}

// Check validates that Connection and Edge types follow Relay naming conventions
func (r *RelayNamingConvention) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all types in the schema
	for _, def := range schema.Types {
		// Skip built-in types and introspection types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		lowerCaseDefName := strings.ToLower(def.Name)

		// Check Connection types
		if strings.HasSuffix(lowerCaseDefName, "connection") {
			errors = append(errors, r.validateConnectionNaming(def, source)...)
		}

		// Check Edge types
		if strings.HasSuffix(lowerCaseDefName, "edge") {
			errors = append(errors, r.validateEdgeNaming(def, source)...)
		}
	}

	return errors
}

// validateConnectionNaming validates that a Connection type follows the [Entity]Connection naming convention
func (r *RelayNamingConvention) validateConnectionNaming(connectionType *ast.Definition, source *ast.Source) []types.LintError {
	var errors []types.LintError

	line, column := 1, 1
	if connectionType.Position != nil {
		line = connectionType.Position.Line
		column = connectionType.Position.Column
	}

	// Rule 1: Connection must be named exactly [Entity]Connection (case-sensitive)
	if !strings.HasSuffix(connectionType.Name, "Connection") {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Connection type `%s` must follow the naming convention [Entity]Connection with proper case.",
				connectionType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
		return errors
	}

	// Rule 2: The entity part must be a valid PascalCase identifier
	entityName := r.extractEntityFromConnection(connectionType.Name)
	if entityName == "" {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Connection type `%s` must have a valid entity name before 'Connection'.",
				connectionType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
		return errors
	} else if !isPascalCase(entityName) {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Connection type `%s` entity name `%s` must be PascalCase.",
				connectionType.Name, entityName),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
		return errors
	}

	// Rule 3: Connection type must have an "edges" field that references the correct Edge type
	edgesField := r.findField(connectionType, "edges")
	if edgesField != nil {
		expectedEdgeTypeName := entityName + "Edge"
		actualEdgeTypeName := r.getEdgeTypeFromEdgesField(edgesField.Type)
		if actualEdgeTypeName != "" && actualEdgeTypeName != expectedEdgeTypeName {
			fieldLine, fieldColumn := line, column
			if edgesField.Position != nil {
				fieldLine = edgesField.Position.Line
				fieldColumn = edgesField.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Connection type `%s` edges field must reference `%s`, but references `%s`.",
					connectionType.Name, expectedEdgeTypeName, actualEdgeTypeName),
				Location: types.Location{
					Line:   fieldLine,
					Column: fieldColumn,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}

// validateEdgeNaming validates that an Edge type follows the [Entity]Edge naming convention
func (r *RelayNamingConvention) validateEdgeNaming(edgeType *ast.Definition, source *ast.Source) []types.LintError {
	var errors []types.LintError

	line, column := 1, 1
	if edgeType.Position != nil {
		line = edgeType.Position.Line
		column = edgeType.Position.Column
	}

	// Rule 1: Edge must be named exactly [Entity]Edge (case-sensitive)
	if !strings.HasSuffix(edgeType.Name, "Edge") {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Edge type `%s` must follow the naming convention [Entity]Edge with proper case.",
				edgeType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
		return errors
	}

	// Rule 2: The entity part must be a valid PascalCase identifier
	entityName := r.extractEntityFromEdge(edgeType.Name)
	if entityName == "" {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Edge type `%s` must have a valid entity name before 'Edge'.",
				edgeType.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	} else if !isPascalCase(entityName) {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Edge type `%s` entity name `%s` must be PascalCase.",
				edgeType.Name, entityName),
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

// extractEntityFromConnection extracts the entity name from a Connection type name
// e.g., "UserConnection" -> "User", "MyEntityConnection" -> "MyEntity"
func (r *RelayNamingConvention) extractEntityFromConnection(connectionName string) string {
	if !strings.HasSuffix(connectionName, "Connection") {
		return ""
	}
	entityName := strings.TrimSuffix(connectionName, "Connection")
	return entityName
}

// extractEntityFromEdge extracts the entity name from an Edge type name
// e.g., "UserEdge" -> "User", "MyEntityEdge" -> "MyEntity"
func (r *RelayNamingConvention) extractEntityFromEdge(edgeName string) string {
	if !strings.HasSuffix(edgeName, "Edge") {
		return ""
	}
	entityName := strings.TrimSuffix(edgeName, "Edge")
	return entityName
}

// findField finds a field by name in a type definition
func (r *RelayNamingConvention) findField(typeDef *ast.Definition, fieldName string) *ast.FieldDefinition {
	for _, field := range typeDef.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

// getEdgeTypeFromEdgesField extracts the Edge type name from an edges field type
// Handles cases like [UserEdge], [UserEdge!], [UserEdge]!, [UserEdge!]!
func (r *RelayNamingConvention) getEdgeTypeFromEdgesField(fieldType *ast.Type) string {
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
