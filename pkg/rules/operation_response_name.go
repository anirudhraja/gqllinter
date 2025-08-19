package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// OperationResponseName checks that query/mutation response types follow the standard naming pattern
type OperationResponseName struct{}

// NewOperationResponseName creates a new instance of the OperationResponseName rule
func NewOperationResponseName() *OperationResponseName {
	return &OperationResponseName{}
}

// Name returns the rule name
func (r *OperationResponseName) Name() string {
	return "operation-response-name"
}

// Description returns what this rule checks
func (r *OperationResponseName) Description() string {
	return "Require query/mutation response type to NOT be called Query/Mutation name + 'Response[Version]' in PascalCase and must be non-nullable"
}

// Check validates that query/mutation response types follow the standard naming pattern
func (r *OperationResponseName) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check Mutation type
	if schema.Mutation != nil {
		errors = append(errors, r.checkFields(schema.Mutation.Fields, "Mutation", source)...)
	}

	// Check Query type
	if schema.Query != nil {
		errors = append(errors, r.checkFields(schema.Query.Fields, "Query", source)...)
	}

	return errors
}

// checkFields validates fields for a given operation type (Query or Mutation)
func (r *OperationResponseName) checkFields(fields ast.FieldList, operationType string, source *ast.Source) []types.LintError {
	var errors []types.LintError

	for _, field := range fields {
		// Skip introspection fields
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		// Check if response type is non-nullable
		//if !r.isNonNullType(field.Type) {
		//	line, column := 1, 1
		//	if field.Position != nil {
		//		line = field.Position.Line
		//		column = field.Position.Column
		//	}
		//
		//	actualType := r.typeToString(field.Type)
		//
		//	errors = append(errors, types.LintError{
		//		Message: fmt.Sprintf("%s `%s` response type should be non-nullable (`%s!` instead of `%s`).", operationType, field.Name, r.getBaseTypeName(field.Type), actualType),
		//		Location: types.Location{
		//			Line:   line,
		//			Column: column,
		//			File:   source.Name,
		//		},
		//		Rule: r.Name(),
		//	})
		//}

		// Check if the response type follows the forbidden naming convention
		forbiddenResponseType := r.capitalizeFirst(field.Name) + "Response"
		actualResponseType := r.getBaseTypeName(field.Type)

		if r.isInvalidResponseType(actualResponseType, forbiddenResponseType) {
			line, column := 1, 1
			if field.Position != nil {
				line = field.Position.Line
				column = field.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("%s `%s` response type should not be named `%s` or `%s[Version]`", operationType, field.Name, forbiddenResponseType, forbiddenResponseType),
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

// getBaseTypeName extracts the base type name from a field type
func (r *OperationResponseName) getBaseTypeName(fieldType *ast.Type) string {
	// Unwrap lists and non-nulls to get the base type
	baseType := fieldType
	for baseType.Elem != nil {
		baseType = baseType.Elem
	}
	return baseType.Name()
}

// isInvalidResponseType checks if the actual type name matches the forbidden pattern
// Forbids both "Response" and "Response[Version]" patterns (e.g., "CreateUserResponse", "CreateUserResponseV2")
func (r *OperationResponseName) isInvalidResponseType(actualType, forbiddenType string) bool {
	// Direct match
	if actualType == forbiddenType {
		return true
	}

	// Check for versioned pattern: {ForbiddenType}V{number} or {ForbiddenType}Version{number}
	if strings.HasPrefix(actualType, forbiddenType) {
		suffix := actualType[len(forbiddenType):]
		// Forbid patterns like "V1", "V2", "Version1", "Version2", etc.
		if strings.HasPrefix(suffix, "V") || strings.HasPrefix(suffix, "Version") {
			return true
		}
	}

	return false
}

// capitalizeFirst capitalizes the first letter of a string
func (r *OperationResponseName) capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// isNonNullType checks if a type is non-null
func (r *OperationResponseName) isNonNullType(fieldType *ast.Type) bool {
	return fieldType.NonNull
}

// typeToString converts an AST type to its string representation
func (r *OperationResponseName) typeToString(fieldType *ast.Type) string {
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
