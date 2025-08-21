package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// RelayArguments checks that fields returning Connection types have proper Relay pagination arguments
type RelayArguments struct{}

// NewRelayArguments creates a new instance of the RelayArguments rule
func NewRelayArguments() *RelayArguments {
	return &RelayArguments{}
}

// Name returns the rule name
func (r *RelayArguments) Name() string {
	return "relay-arguments"
}

// Description returns what this rule checks
func (r *RelayArguments) Description() string {
	return "Ensure fields returning Connection types include proper Relay pagination arguments (first/after for forward, last/before for backward)"
}

// Check validates that fields returning Connection types have proper pagination arguments
func (r *RelayArguments) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find all Connection types in the schema
	connectionTypes := make(map[string]bool)
	for _, def := range schema.Types {
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}
		lowerCaseDefName := strings.ToLower(def.Name)
		if strings.HasSuffix(lowerCaseDefName, "connection") {
			connectionTypes[def.Name] = true
		}
	}

	// Check all types in the schema for fields that return Connection types
	for _, def := range schema.Types {
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		// Only check Object types and Interface types that can have fields
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			for _, field := range def.Fields {
				connectionTypeName := r.getConnectionTypeFromField(field)
				if connectionTypeName != "" && connectionTypes[connectionTypeName] {
					errors = append(errors, r.validateConnectionFieldArguments(def, field, source)...)
				}
			}
		}
	}

	return errors
}

// getConnectionTypeFromField extracts the Connection type name from a field's return type
func (r *RelayArguments) getConnectionTypeFromField(field *ast.FieldDefinition) string {
	fieldType := field.Type

	// Navigate through NonNull and List wrappers to get the base type
	for fieldType != nil {
		if fieldType.NamedType != "" {
			return fieldType.NamedType
		}
		fieldType = fieldType.Elem
	}

	return ""
}

// validateConnectionFieldArguments validates that a field returning a Connection type has proper pagination arguments
func (r *RelayArguments) validateConnectionFieldArguments(parentType *ast.Definition, field *ast.FieldDefinition, source *ast.Source) []types.LintError {
	var errors []types.LintError

	line, column := 1, 1
	if field.Position != nil {
		line = field.Position.Line
		column = field.Position.Column
	}

	// Check for arguments presence (regardless of type)
	hasFirstArg := r.hasArgument(field, "first")
	hasAfterArg := r.hasArgument(field, "after")
	hasLastArg := r.hasArgument(field, "last")
	hasBeforeArg := r.hasArgument(field, "before")

	// Check for partial implementations first
	hasPartialForward := hasFirstArg || hasAfterArg
	hasPartialBackward := hasLastArg || hasBeforeArg

	// Check individual argument issues for partial implementations
	if hasFirstArg && !hasAfterArg {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Field `%s.%s` has `first` argument but is missing `after` argument for complete forward pagination.",
				parentType.Name, field.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	if hasAfterArg && !hasFirstArg {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Field `%s.%s` has `after` argument but is missing `first` argument for complete forward pagination.",
				parentType.Name, field.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	if hasLastArg && !hasBeforeArg {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Field `%s.%s` has `last` argument but is missing `before` argument for complete backward pagination.",
				parentType.Name, field.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	if hasBeforeArg && !hasLastArg {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Field `%s.%s` has `before` argument but is missing `last` argument for complete backward pagination.",
				parentType.Name, field.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	// If no complete pagination and no partial implementations, report the general error
	if !hasPartialForward && !hasPartialBackward {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Field `%s.%s` returns Connection type but lacks proper pagination arguments. Must include forward pagination arguments (first and after), backward pagination arguments (last and before), or both.",
				parentType.Name, field.Name),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	// Check individual argument type validations
	errors = append(errors, r.validateArgumentTypes(parentType, field, source)...)

	return errors
}

// hasValidFirstArgument checks if the field has a valid 'first' argument (non-negative integer)
func (r *RelayArguments) hasValidFirstArgument(field *ast.FieldDefinition) bool {
	arg := r.findArgument(field, "first")
	return arg != nil && r.isIntegerType(arg.Type)
}

// hasValidAfterArgument checks if the field has a valid 'after' argument (Cursor type)
func (r *RelayArguments) hasValidAfterArgument(field *ast.FieldDefinition) bool {
	arg := r.findArgument(field, "after")
	return arg != nil && r.isCursorType(arg.Type)
}

// hasValidLastArgument checks if the field has a valid 'last' argument (non-negative integer)
func (r *RelayArguments) hasValidLastArgument(field *ast.FieldDefinition) bool {
	arg := r.findArgument(field, "last")
	return arg != nil && r.isIntegerType(arg.Type)
}

// hasValidBeforeArgument checks if the field has a valid 'before' argument (Cursor type)
func (r *RelayArguments) hasValidBeforeArgument(field *ast.FieldDefinition) bool {
	arg := r.findArgument(field, "before")
	return arg != nil && r.isCursorType(arg.Type)
}

// hasArgument checks if the field has an argument with the given name (regardless of type)
func (r *RelayArguments) hasArgument(field *ast.FieldDefinition, argName string) bool {
	return r.findArgument(field, argName) != nil
}

// validateArgumentTypes validates the types of pagination arguments
func (r *RelayArguments) validateArgumentTypes(parentType *ast.Definition, field *ast.FieldDefinition, source *ast.Source) []types.LintError {
	var errors []types.LintError

	line, column := 1, 1
	if field.Position != nil {
		line = field.Position.Line
		column = field.Position.Column
	}

	// Check 'first' argument type
	if firstArg := r.findArgument(field, "first"); firstArg != nil {
		if !r.isIntegerType(firstArg.Type) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Field `%s.%s` argument `first` must be a non-negative integer type (Int), but is %s.",
					parentType.Name, field.Name, r.typeToString(firstArg.Type)),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	// Check 'after' argument type
	if afterArg := r.findArgument(field, "after"); afterArg != nil {
		if !r.isCursorType(afterArg.Type) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Field `%s.%s` argument `after` must be a Cursor type (String), but is %s.",
					parentType.Name, field.Name, r.typeToString(afterArg.Type)),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	// Check 'last' argument type
	if lastArg := r.findArgument(field, "last"); lastArg != nil {
		if !r.isIntegerType(lastArg.Type) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Field `%s.%s` argument `last` must be a non-negative integer type (Int), but is %s.",
					parentType.Name, field.Name, r.typeToString(lastArg.Type)),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	// Check 'before' argument type
	if beforeArg := r.findArgument(field, "before"); beforeArg != nil {
		if !r.isCursorType(beforeArg.Type) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Field `%s.%s` argument `before` must be a Cursor type (String), but is %s.",
					parentType.Name, field.Name, r.typeToString(beforeArg.Type)),
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

// findArgument finds an argument by name in a field definition
func (r *RelayArguments) findArgument(field *ast.FieldDefinition, argName string) *ast.ArgumentDefinition {
	for _, arg := range field.Arguments {
		if arg.Name == argName {
			return arg
		}
	}
	return nil
}

// isIntegerType checks if a type is an integer type (Int with optional NonNull)
func (r *RelayArguments) isIntegerType(argType *ast.Type) bool {
	// Check if it's directly Int
	if argType.NamedType == "Int" {
		return true
	}

	// Check if it's NonNull Int
	if argType.NonNull && argType.Elem != nil && argType.Elem.NamedType == "Int" {
		return true
	}

	return false
}

// isCursorType checks if a type is a Cursor type (String with optional NonNull)
func (r *RelayArguments) isCursorType(argType *ast.Type) bool {
	// Check if it's directly String (Cursor is typically a String)
	if argType.NamedType == "String" {
		return true
	}

	// Check if it's NonNull String
	if argType.NonNull && argType.Elem != nil && argType.Elem.NamedType == "String" {
		return true
	}

	// Also accept explicit Cursor type if defined
	if argType.NamedType == "Cursor" {
		return true
	}

	// Check if it's NonNull Cursor
	if argType.NonNull && argType.Elem != nil && argType.Elem.NamedType == "Cursor" {
		return true
	}

	return false
}

// typeToString converts a GraphQL type to its string representation
func (r *RelayArguments) typeToString(fieldType *ast.Type) string {
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
