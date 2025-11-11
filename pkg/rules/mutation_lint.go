package rules

import (
	"fmt"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// MutationLint validates mutation response union patterns
type MutationLint struct{}

// NewMutationLint creates a new instance of the MutationLint rule
func NewMutationLint() *MutationLint {
	return &MutationLint{}
}

// Name returns the rule name
func (r *MutationLint) Name() string {
	return "mutation-lint"
}

// Description returns what this rule checks
func (r *MutationLint) Description() string {
	return "Validates that mutations return @responseUnion unions, @error types are only in mutation/query unions, unions have exactly one success type, and all other types are @error types"
}

// Check validates mutation response union rules
func (r *MutationLint) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check that every mutation returns a @responseUnion union
	errors = append(errors, r.validateMutationResponseUnions(schema, source)...)

	// Check that @error types are only in mutation and query unions
	errors = append(errors, r.validateErrorTypeUsage(schema, source)...)

	// Check that @responseUnion unions have exactly one success type
	errors = append(errors, r.validateUnionSuccessTypes(schema, source)...)

	// Check that non-success types in @responseUnion unions are @error types
	errors = append(errors, r.validateUnionErrorTypes(schema, source)...)

	return errors
}

// validateMutationResponseUnions checks that every mutation returns a @responseUnion union
func (r *MutationLint) validateMutationResponseUnions(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find the Mutation type
	mutationType := schema.Types["Mutation"]
	if mutationType == nil {
		return errors
	}

	// Check each mutation field
	for _, field := range mutationType.Fields {
		returnTypeName := field.Type.NamedType
		if returnTypeName == "" {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Mutation field '%s' must return a union type with @responseUnion directive, but returns a list type", field.Name),
				Location: types.Location{
					Line:   field.Position.Line,
					Column: field.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
			continue
		}
		returnType := schema.Types[returnTypeName]

		// Check if the return type is a union with @responseUnion directive
		if returnType.Kind != ast.Union {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Mutation field '%s' must return a union type with @responseUnion directive, but returns '%s'", field.Name, returnTypeName),
				Location: types.Location{
					Line:   field.Position.Line,
					Column: field.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
			continue
		}

		// Check if the union has @responseUnion directive
		if !r.hasResponseUnionDirective(returnType) {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Mutation field '%s' returns union '%s' which must have @responseUnion directive", field.Name, returnTypeName),
				Location: types.Location{
					Line:   field.Position.Line,
					Column: field.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}

// validateErrorTypeUsage checks that @error types are only used in mutation and query unions
func (r *MutationLint) validateErrorTypeUsage(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find all @error types
	errorTypes := r.findErrorTypes(schema)

	// Check each @error type usage
	for _, errorTypeName := range errorTypes {
		errors = append(errors, r.validateErrorTypeOnlyInMutationQueryUnions(schema, source, errorTypeName)...)
	}

	return errors
}

// validateUnionSuccessTypes checks that @responseUnion unions have exactly one success type
func (r *MutationLint) validateUnionSuccessTypes(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find all @responseUnion unions
	responseUnions := r.findResponseUnions(schema)

	for _, unionType := range responseUnions {
		successTypes := []string{}

		// Categorize union member types
		for _, memberType := range unionType.Types {
			if !r.hasErrorDirective(schema.Types[memberType]) {
				successTypes = append(successTypes, memberType)
			}
		}

		// Check for exactly one success type
		if len(successTypes) == 0 {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Union '%s' with @responseUnion directive must have exactly one success type (non-@error type), but has none", unionType.Name),
				Location: types.Location{
					Line:   unionType.Position.Line,
					Column: unionType.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		} else if len(successTypes) > 1 {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Union '%s' with @responseUnion directive must have exactly one success type (non-@error type), but has %d: %v", unionType.Name, len(successTypes), successTypes),
				Location: types.Location{
					Line:   unionType.Position.Line,
					Column: unionType.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}

// validateUnionErrorTypes checks that non-success types in @responseUnion unions are @error types
func (r *MutationLint) validateUnionErrorTypes(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find all @responseUnion unions
	responseUnions := r.findResponseUnions(schema)

	for _, unionType := range responseUnions {
		successTypes := []string{}

		// First pass: count success types
		for _, memberTypeName := range unionType.Types {
			memberType := schema.Types[memberTypeName]
			if memberType != nil && !r.hasErrorDirective(memberType) {
				successTypes = append(successTypes, memberTypeName)
			}
		}

		// Only validate error types if we have exactly one success type
		// (multiple success types will be caught by validateUnionSuccessTypes)
		if len(successTypes) == 1 {
			successTypeName := successTypes[0]

			// Check each member type that is not the success type
			for _, memberTypeName := range unionType.Types {
				if memberTypeName == successTypeName {
					continue // This is the allowed success type
				}

				memberType := schema.Types[memberTypeName]
				if memberType != nil && !r.hasErrorDirective(memberType) {
					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Union '%s' with @responseUnion directive contains type '%s' which is not an @error type. All types except the single success type must have @error directive", unionType.Name, memberTypeName),
						Location: types.Location{
							Line:   unionType.Position.Line,
							Column: unionType.Position.Column,
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

// hasResponseUnionDirective checks if a type has the @responseUnion directive
func (r *MutationLint) hasResponseUnionDirective(typeDefinition *ast.Definition) bool {
	if typeDefinition == nil {
		return false
	}
	for _, directive := range typeDefinition.Directives {
		if directive.Name == "responseUnion" {
			return true
		}
	}
	return false
}

// hasErrorDirective checks if a type has the @error directive
func (r *MutationLint) hasErrorDirective(typeDefinition *ast.Definition) bool {
	if typeDefinition == nil {
		return false
	}
	for _, directive := range typeDefinition.Directives {
		if directive.Name == "error" {
			return true
		}
	}
	return false
}

// findErrorTypes returns all type names that have @error directive
func (r *MutationLint) findErrorTypes(schema *ast.Schema) []string {
	var errorTypes []string
	for _, typeDef := range schema.Types {
		if r.hasErrorDirective(typeDef) {
			errorTypes = append(errorTypes, typeDef.Name)
		}
	}
	return errorTypes
}

// findResponseUnions returns all union types that have @responseUnion directive
func (r *MutationLint) findResponseUnions(schema *ast.Schema) []*ast.Definition {
	var responseUnions []*ast.Definition
	for _, typeDef := range schema.Types {
		if typeDef.Kind == ast.Union && r.hasResponseUnionDirective(typeDef) {
			responseUnions = append(responseUnions, typeDef)
		}
	}
	return responseUnions
}

// validateErrorTypeOnlyInMutationQueryUnions checks that @error types are only used in mutation/query unions
func (r *MutationLint) validateErrorTypeOnlyInMutationQueryUnions(schema *ast.Schema, source *ast.Source, errorTypeName string) []types.LintError {
	var errors []types.LintError

	// Find all unions that contain this error type
	for _, typeDef := range schema.Types {
		if typeDef.Kind != ast.Union {
			continue
		}

		// Check if this union contains the error type
		containsErrorType := false
		for _, memberType := range typeDef.Types {
			if memberType == errorTypeName {
				containsErrorType = true
				break
			}
		}

		if !containsErrorType {
			continue
		}

		// Check if this union is used in mutation or query responses
		isUsedInMutationOrQuery := r.isUnionUsedInMutationOrQuery(schema, typeDef.Name)

		if !isUsedInMutationOrQuery {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Type '%s' has @error directive but is used in union '%s' which is not returned by any mutation or query. @error types should only be part of mutation and query response unions", errorTypeName, typeDef.Name),
				Location: types.Location{
					Line:   typeDef.Position.Line,
					Column: typeDef.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}

// isUnionUsedInMutationOrQuery checks if a union type is used as return type in mutation or query
func (r *MutationLint) isUnionUsedInMutationOrQuery(schema *ast.Schema, unionTypeName string) bool {
	// Check Mutation type
	mutationType := schema.Types["Mutation"]
	if mutationType != nil {
		for _, field := range mutationType.Fields {
			if field.Type.NamedType == unionTypeName {
				return true
			}
		}
	}

	// Check Query type
	queryType := schema.Types["Query"]
	if queryType != nil {
		for _, field := range queryType.Fields {
			if field.Type.NamedType == unionTypeName {
				return true
			}
		}
	}

	return false
}
