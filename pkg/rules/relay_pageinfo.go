package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// RelayPageInfo checks that PageInfo objects comply with the Relay specification
type RelayPageInfo struct{}

// NewRelayPageInfo creates a new instance of the RelayPageInfo rule
func NewRelayPageInfo() *RelayPageInfo {
	return &RelayPageInfo{}
}

// Name returns the rule name
func (r *RelayPageInfo) Name() string {
	return "relay-pageinfo"
}

// Description returns what this rule checks
func (r *RelayPageInfo) Description() string {
	return "Ensure PageInfo objects comply with the Relay specification requirements"
}

// Check validates that PageInfo objects follow Relay specifications
func (r *RelayPageInfo) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all object types
	for _, def := range schema.Types {
		if def.Kind == ast.Object {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			// Check if this is a PageInfo type
			if def.Name == "PageInfo" {
				errors = append(errors, r.validatePageInfoType(def, source)...)
			}
		}
	}

	return errors
}

// validatePageInfoType validates that a PageInfo type meets Relay specifications
func (r *RelayPageInfo) validatePageInfoType(pageInfoType *ast.Definition, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Required fields according to Relay spec
	requiredFields := map[string]RequiredField{
		"hasNextPage": {
			name:         "hasNextPage",
			expectedType: "Boolean!",
			description:  "indicates whether more edges exist following the current page",
		},
		"hasPreviousPage": {
			name:         "hasPreviousPage", 
			expectedType: "Boolean!",
			description:  "indicates whether more edges exist prior to the current page",
		},
		"startCursor": {
			name:         "startCursor",
			expectedType: "String",
			description:  "cursor corresponding to the first edge in the current page (nullable if no results)",
		},
		"endCursor": {
			name:         "endCursor",
			expectedType: "String", 
			description:  "cursor corresponding to the last edge in the current page (nullable if no results)",
		},
	}

	// Check each required field
	for fieldName, required := range requiredFields {
		field := r.findField(pageInfoType, fieldName)
		if field == nil {
			line, column := 1, 1
			if pageInfoType.Position != nil {
				line = pageInfoType.Position.Line
				column = pageInfoType.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("PageInfo must contain field `%s` that returns %s (%s).", 
					required.name, required.expectedType, required.description),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
			continue
		}

		// Validate field type
		actualType := r.typeToString(field.Type)
		if actualType != required.expectedType {
			line, column := 1, 1
			if field.Position != nil {
				line = field.Position.Line
				column = field.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("PageInfo field `%s` must return %s, but returns %s.", 
					required.name, required.expectedType, actualType),
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

// RequiredField represents a required field in PageInfo
type RequiredField struct {
	name         string
	expectedType string
	description  string
}

// findField finds a field by name in a type definition
func (r *RelayPageInfo) findField(typeDef *ast.Definition, fieldName string) *ast.FieldDefinition {
	for _, field := range typeDef.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

// typeToString converts a GraphQL type to its string representation
func (r *RelayPageInfo) typeToString(fieldType *ast.Type) string {
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