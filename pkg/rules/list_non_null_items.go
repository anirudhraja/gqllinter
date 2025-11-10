package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// ListNonNullItems checks that list types contain non-null items (recursively)
type ListNonNullItems struct{}

// NewListNonNullItems creates a new instance of the ListNonNullItems rule
func NewListNonNullItems() *ListNonNullItems {
	return &ListNonNullItems{}
}

// Name returns the rule name
func (r *ListNonNullItems) Name() string {
	return "list-non-null-items"
}

// Description returns what this rule checks
func (r *ListNonNullItems) Description() string {
	return "Requires list being returned to not contain null values (checks recursively for nested lists)"
}

// Check validates that list fields contain non-null types
func (r *ListNonNullItems) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check fields in object types and interfaces
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.InputObject {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") || r.isConnectionType(def.Name) {
				continue
			}

			for _, field := range def.Fields {
				// Skip built-in fields and introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}

				if r.isListWithNullableItems(field.Type) {
					line, column := 1, 1
					if field.Position != nil {
						line = field.Position.Line
						column = field.Position.Column
					}

					suggestion := r.suggestNonNullVariant(field.Type)

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("List field `%s.%s` contains nullable items. Use `%s` instead to prevent null pointer issues.", def.Name, field.Name, suggestion),
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

// isListWithNullableItems checks if a type is a list containing nullable items (recursively)
func (r *ListNonNullItems) isListWithNullableItems(fieldType *ast.Type) bool {
	if fieldType.Elem != nil {
		return r.checkTypeRecursively(fieldType.Elem)
	}
	return false
}

// checkTypeRecursively checks if any list at any nesting level contains nullable items
func (r *ListNonNullItems) checkTypeRecursively(fieldType *ast.Type) bool {
	if fieldType == nil {
		return false
	}

	if !fieldType.NonNull {
		return true
	}

	if fieldType.NamedType == "" && fieldType.Elem != nil {
		return r.checkTypeRecursively(fieldType.Elem)
	}

	return !fieldType.NonNull
}

// suggestNonNullVariant suggests the non-null variant of a list type (recursively)
func (r *ListNonNullItems) suggestNonNullVariant(fieldType *ast.Type) string {
	correctedType := r.makeListItemsNonNull(fieldType.Elem)
	correctedType = &ast.Type{
		NonNull: fieldType.NonNull,
		Elem:    correctedType,
	}
	return correctedType.String()
}

// makeListItemsNonNull creates a copy of the type with all list items marked as non-null
func (r *ListNonNullItems) makeListItemsNonNull(fieldType *ast.Type) *ast.Type {
	// Handle the outer non-null wrapper
	if fieldType.Elem != nil {
		return &ast.Type{
			NonNull: true,
			Elem:    r.makeListItemsNonNull(fieldType.Elem),
		}
	}

	fieldType.NonNull = true

	// Base case: named type
	return fieldType
}

// isConnectionType checks if a type name indicates a connection type
func (r *ListNonNullItems) isConnectionType(typeName string) bool {
	return strings.HasSuffix(strings.ToLower(typeName), "connection")
}
