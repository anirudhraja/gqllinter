package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// ListNonNullItems checks that list types contain non-null items
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
	return "List types should contain non-null items to prevent null pointer issues and improve type safety"
}

// Check validates that list fields contain non-null types
func (r *ListNonNullItems) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check fields in object types and interfaces
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.InputObject {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
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

// isListWithNullableItems checks if a type is a list containing nullable items
func (r *ListNonNullItems) isListWithNullableItems(fieldType *ast.Type) bool {
	// Skip non-null wrapper to get to the actual type
	currentType := fieldType
	if currentType.NonNull && currentType.Elem != nil {
		currentType = currentType.Elem
	}

	// Check if this is a list type (no NamedType means it's a wrapper type like List)
	if currentType.NamedType == "" && currentType.Elem != nil {
		// This is a list type, check if the inner type is nullable
		innerType := currentType.Elem

		// Inner type is nullable if it's not marked as NonNull
		return !innerType.NonNull
	}

	return false
}

// suggestNonNullVariant suggests the non-null variant of a list type
func (r *ListNonNullItems) suggestNonNullVariant(fieldType *ast.Type) string {
	typeStr := r.typeToString(fieldType)

	// Convert [Type] to [Type!] or [Type]! to [Type!]!
	if strings.Contains(typeStr, "[") && strings.Contains(typeStr, "]") {
		// Find the inner type and make it non-null
		start := strings.Index(typeStr, "[")
		end := strings.LastIndex(typeStr, "]")

		if start != -1 && end != -1 {
			innerType := typeStr[start+1 : end]

			// Add ! if not already present
			if !strings.HasSuffix(innerType, "!") {
				innerType += "!"
			}

			result := typeStr[:start+1] + innerType + typeStr[end:]
			return result
		}
	}

	return typeStr
}

// typeToString converts an AST type to its string representation
func (r *ListNonNullItems) typeToString(fieldType *ast.Type) string {
	if fieldType.NamedType != "" {
		if fieldType.NonNull {
			return fieldType.NamedType + "!"
		}
		return fieldType.NamedType
	}

	if fieldType.Elem != nil {
		innerStr := r.typeToString(fieldType.Elem)
		if fieldType.NonNull {
			return "[" + innerStr + "]!"
		}
		return "[" + innerStr + "]"
	}

	return "Unknown"
}
