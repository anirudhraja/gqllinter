package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// InputName checks that query/mutation arguments follow the standard naming pattern
type InputName struct{}

// NewInputName creates a new instance of the InputName rule
func NewInputName() *InputName {
	return &InputName{}
}

// Name returns the rule name
func (r *InputName) Name() string {
	return "operation-input-name"
}

// Description returns what this rule checks
func (r *InputName) Description() string {
	return "Require query/mutation argument to be always called 'request' and input type to NOT be called Query/Mutation name + 'Request[Version]' in PascalCase"
}

// Check validates that query/mutation arguments follow the standard naming pattern
func (r *InputName) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
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
func (r *InputName) checkFields(fields ast.FieldList, operationType string, source *ast.Source) []types.LintError {
	var errors []types.LintError

	for _, field := range fields {
		// Skip introspection fields
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		// Check field arguments
		if len(field.Arguments) > 0 {
			// For operations with arguments, we expect a single "request" argument
			if len(field.Arguments) == 1 {
				arg := field.Arguments[0]

				// Check if the argument is named "request"
				if arg.Name != "input" {
					line, column := 1, 1
					if arg.Position != nil {
						line = arg.Position.Line
						column = arg.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("%s `%s` argument should be named 'input', not '%s'.", operationType, field.Name, arg.Name),
						Location: types.Location{
							Line:   line,
							Column: column,
							File:   source.Name,
						},
						Rule: r.Name(),
					})
				}

				// Check if the input type follows the naming convention
				expectedInputType := r.capitalizeFirst(field.Name) + "Request"
				actualInputType := r.getBaseTypeName(arg.Type)

				if r.isInValidRequestType(actualInputType, expectedInputType) {
					line, column := 1, 1
					if arg.Position != nil {
						line = arg.Position.Line
						column = arg.Position.Column
					}

					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("%s `%s` input type should not be named `%s` or `%s[Version]`", operationType, field.Name, expectedInputType, expectedInputType),
						Location: types.Location{
							Line:   line,
							Column: column,
							File:   source.Name,
						},
						Rule: r.Name(),
					})
				}
			} else if len(field.Arguments) > 1 {
				// Multiple arguments - suggest consolidating into a request type
				line, column := 1, 1
				if field.Position != nil {
					line = field.Position.Line
					column = field.Position.Column
				}

				//expectedInputType := r.capitalizeFirst(field.Name) + "Request"
				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("%s `%s` has %d arguments. Consider consolidating into a single 'input' argument of a properly named input type (not %sRequest).", operationType, field.Name, len(field.Arguments), r.capitalizeFirst(field.Name)),
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

	return errors
}

// getBaseTypeName extracts the base type name from a field type
func (r *InputName) getBaseTypeName(fieldType *ast.Type) string {
	// Unwrap lists and non-nulls to get the base type
	baseType := fieldType
	for baseType.Elem != nil {
		baseType = baseType.Elem
	}
	return baseType.Name()
}

// isValidRequestType checks if the actual type name matches the expected pattern
// Allows for both "Request" and "Request[Version]" patterns (e.g., "CreateUserRequest", "CreateUserRequestV2")
func (r *InputName) isInValidRequestType(actualType, expectedType string) bool {
	// Direct match
	if actualType == expectedType {
		return true
	}

	// Check for versioned pattern: {ExpectedType}V{number} or {ExpectedType}Version{number}
	if strings.HasPrefix(actualType, expectedType) {
		suffix := actualType[len(expectedType):]
		// Allow patterns like "V1", "V2", "Version1", "Version2", etc.
		if strings.HasPrefix(suffix, "V") || strings.HasPrefix(suffix, "Version") {
			return true
		}
	}

	return false
}

// capitalizeFirst capitalizes the first letter of a string
func (r *InputName) capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
