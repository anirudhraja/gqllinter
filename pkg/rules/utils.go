package rules

import "github.com/nishant-rn/gqlparser/v2/ast"

// isNestedListType checks if a type is a nested list (list of lists)
func isNestedListType(fieldType *ast.Type) bool {
	// First, check if this is a list
	if !isListType(fieldType) {
		return false
	}

	// Get the element type of the list
	elementType := getListElementType(fieldType)
	if elementType == nil {
		return false
	}

	// Check if the element type is also a list
	return isListType(elementType)
}

// isListType checks if a type is a list type (with or without NonNull wrapper)
func isListType(fieldType *ast.Type) bool {
	// Check if it's directly a list
	if fieldType.Elem != nil && fieldType.NamedType == "" {
		return true
	}

	// Check if it's a NonNull wrapper around a list
	if fieldType.NonNull && fieldType.Elem != nil {
		return isListType(fieldType.Elem)
	}

	return false
}

// getListElementType gets the element type of a list, handling NonNull wrappers
func getListElementType(fieldType *ast.Type) *ast.Type {
	// Navigate through the type structure to find the first list and return its element
	current := fieldType

	// Keep going until we find a list type (NamedType == "" and Elem != nil)
	for current != nil {
		// If this is a list type, return its element
		if current.NamedType == "" && current.Elem != nil {
			return current.Elem
		}

		// If this is a NonNull wrapper, continue to the wrapped type
		if current.NonNull && current.Elem != nil {
			current = current.Elem
		} else {
			break
		}
	}

	return nil
}
