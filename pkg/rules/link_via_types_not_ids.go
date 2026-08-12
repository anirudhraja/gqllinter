package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// LinkViaTypesNotIds checks that fields reference entity types directly instead of storing IDs
type LinkViaTypesNotIds struct{}

// NewLinkViaTypesNotIds creates a new instance of the LinkViaTypesNotIds rule
func NewLinkViaTypesNotIds() *LinkViaTypesNotIds {
	return &LinkViaTypesNotIds{}
}

// Name returns the rule name
func (r *LinkViaTypesNotIds) Name() string {
	return "link-via-types-not-ids"
}

// Description returns what this rule checks
func (r *LinkViaTypesNotIds) Description() string {
	return "Fields should reference entity types directly instead of storing IDs of those entities"
}

// Check validates that fields reference types instead of IDs
func (r *LinkViaTypesNotIds) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Build a map of entity types (types with @key directive)
	entityTypes := r.getEntityTypes(schema)

	// Check all object types
	for _, def := range schema.Types {
		// Skip introspection types, built-in types, and non-object types
		if strings.HasPrefix(def.Name, "__") || def.BuiltIn {
			continue
		}

		if def.Kind != ast.Object && def.Kind != ast.Interface {
			continue
		}

		// Check each field in the type
		for _, field := range def.Fields {
			// Skip introspection fields
			if strings.HasPrefix(field.Name, "__") {
				continue
			}

			// Check if this field looks like it's storing an ID reference
			if idFieldError := r.checkFieldForIdReference(field, def.Name, entityTypes, source); idFieldError != nil {
				errors = append(errors, *idFieldError)
			}
		}
	}

	return errors
}

// getEntityTypes returns a map of entity type names (types with @key directive)
func (r *LinkViaTypesNotIds) getEntityTypes(schema *ast.Schema) map[string]bool {
	entityTypes := make(map[string]bool)

	for _, def := range schema.Types {
		// Skip built-in types
		if def.BuiltIn {
			continue
		}

		// Check if type has @key directive
		for _, directive := range def.Directives {
			if directive.Name == "key" {
				entityTypes[def.Name] = true
				break
			}
		}
	}

	return entityTypes
}

// checkFieldForIdReference checks if a field is storing an ID reference instead of a type reference
func (r *LinkViaTypesNotIds) checkFieldForIdReference(field *ast.FieldDefinition, parentTypeName string, entityTypes map[string]bool, source *ast.Source) *types.LintError {
	// Get the base type name (unwrap NonNull and List wrappers)
	baseTypeName := r.getBaseTypeName(field.Type)

	// Check if field ends with "Id" or "ID"
	var possibleEntityName string
	if strings.HasSuffix(field.Name, "Id") && len(field.Name) > 2 {
		possibleEntityName = field.Name[:len(field.Name)-2]
	} else if strings.HasSuffix(field.Name, "ID") && len(field.Name) > 2 {
		possibleEntityName = field.Name[:len(field.Name)-2]
	} else {
		// Field doesn't end with Id/ID, so it's fine
		return nil
	}

	// Check if the field type is a scalar type (String, ID, Int, etc.)
	if !r.isScalarType(baseTypeName) {
		// Field is already using a type reference, which is good
		return nil
	}

	// Convert possibleEntityName to PascalCase to match type naming conventions
	pascalCaseEntityName := r.toPascalCase(possibleEntityName)

	// Check if there's an entity type with the matching name
	if entityTypes[pascalCaseEntityName] {
		line, column := 1, 1
		if field.Position != nil {
			line = field.Position.Line
			column = field.Position.Column
		}

		suggestion := r.makeSuggestion(possibleEntityName, pascalCaseEntityName, field.Type)

		return &types.LintError{
			Message: fmt.Sprintf("Field `%s.%s` should reference the `%s` type directly instead of storing its ID. Consider using `%s` instead of `%s: %s`",
				parentTypeName, field.Name, pascalCaseEntityName, suggestion, field.Name, field.Type.String()),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		}
	}

	return nil
}

// getBaseTypeName returns the base type name, unwrapping NonNull and List wrappers
func (r *LinkViaTypesNotIds) getBaseTypeName(fieldType *ast.Type) string {
	if fieldType.NamedType != "" {
		return fieldType.NamedType
	}
	if fieldType.Elem != nil {
		return r.getBaseTypeName(fieldType.Elem)
	}
	return ""
}

// isScalarType checks if a type name is a scalar type
func (r *LinkViaTypesNotIds) isScalarType(typeName string) bool {
	scalarTypes := map[string]bool{
		"ID":      true,
		"String":  true,
		"Int":     true,
		"Float":   true,
		"Boolean": true,
	}
	return scalarTypes[typeName]
}

// toPascalCase converts a string to PascalCase
func (r *LinkViaTypesNotIds) toPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Convert first character to uppercase
	return strings.ToUpper(string(s[0])) + s[1:]
}

// makeSuggestion creates a suggestion for the field name and type
func (r *LinkViaTypesNotIds) makeSuggestion(fieldPrefix string, entityTypeName string, originalType *ast.Type) string {
	// Convert field name to camelCase (first letter lowercase)
	suggestedFieldName := strings.ToLower(string(fieldPrefix[0])) + fieldPrefix[1:]

	// Preserve nullability from the original type
	if originalType.NonNull {
		return fmt.Sprintf("%s: %s!", suggestedFieldName, entityTypeName)
	}
	return fmt.Sprintf("%s: %s", suggestedFieldName, entityTypeName)
}
