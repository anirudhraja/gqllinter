package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

// KeyDirectivesLint checks @key directive validation rules
type KeyDirectivesLint struct{}

// NewKeyDirectivesLint creates a new instance of the KeyDirectivesLint rule
func NewKeyDirectivesLint() *KeyDirectivesLint {
	return &KeyDirectivesLint{}
}

// Name returns the rule name
func (r *KeyDirectivesLint) Name() string {
	return "key-directive-lint"
}

// Description returns what this rule checks
func (r *KeyDirectivesLint) Description() string {
	return "Validates that all fields specified in @key directive exist in the object type, are primitive/scalar types only, are space-separated (not comma-separated), enforces resolvable: false @key directive constraints, and ensures all fields are included in resolvable: false keys"
}

// Check validates @key directive rules
func (r *KeyDirectivesLint) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all object types for @key directive validation
	errors = append(errors, r.validateKeyDirectiveFields(schema, source)...)

	// Check resolvable: false @key directive constraints
	errors = append(errors, r.validateResolvableFalseConstraints(schema, source)...)

	// Check that resolvable: false keys include all fields
	errors = append(errors, r.validateResolvableFalseIncludesAllFields(schema, source)...)

	return errors
}

// validateKeyDirectiveFields checks that all fields in @key directive exist in the object
func (r *KeyDirectivesLint) validateKeyDirectiveFields(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all object types in the schema
	for _, def := range schema.Types {
		// Only check object types (not interfaces, enums, etc.)
		if def.Kind != ast.Object {
			continue
		}

		// Skip built-in types and introspection types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		// Check for @key directive on the type
		for _, directive := range def.Directives {
			if directive.Name == "key" {
				keyFieldErrors := r.validateKeyFields(def, directive, source, schema)
				errors = append(errors, keyFieldErrors...)
			}
		}
	}

	return errors
}

// validateKeyFields validates that all fields specified in a @key directive exist in the object
func (r *KeyDirectivesLint) validateKeyFields(objectDef *ast.Definition, keyDirective *ast.Directive, source *ast.Source, schema *ast.Schema) []types.LintError {
	var errors []types.LintError

	// Get the fields argument from @key directive
	var fieldsArg *ast.Argument
	for _, arg := range keyDirective.Arguments {
		if arg.Name == "fields" {
			fieldsArg = arg
			break
		}
	}

	// If no fields argument found, skip validation
	if fieldsArg == nil {
		return errors
	}

	// Extract the fields string from the argument value
	fieldsString := r.extractFieldsString(fieldsArg.Value)
	if fieldsString == "" {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Missing or invalid 'fields' argument in @key directive for object '%s'", objectDef.Name),
			Location: types.Location{
				Line:   fieldsArg.Position.Line,
				Column: fieldsArg.Position.Column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
		return errors
	}

	// Check for comma-separated fields (not allowed)
	if r.hasCommaSeparatedFields(fieldsString) {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("@key directive fields must be space-separated, not comma-separated. Found comma in fields: '%s' for object '%s'", fieldsString, objectDef.Name),
			Location: types.Location{
				Line:   fieldsArg.Position.Line,
				Column: fieldsArg.Position.Column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
		return errors
	}

	query := fmt.Sprintf("fragment x on %s { %s }", objectDef.Name, fieldsString)
	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Failed to parse fields in @key directive for object '%s': %v", objectDef.Name, err),
			Location: types.Location{
				Line:   fieldsArg.Position.Line,
				Column: fieldsArg.Position.Column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	selectionSet := doc.Fragments[0].SelectionSet
	for _, sel := range selectionSet {
		fieldSel, ok := sel.(*ast.Field)
		if !ok {
			continue
		}
		fieldName := fieldSel.Name
		line, column := 1, 1
		if keyDirective.Position != nil {
			line = keyDirective.Position.Line
			column = keyDirective.Position.Column
		}
		field := objectDef.Fields.ForName(fieldName)
		if field == nil {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Field '%s' specified in @key directive does not exist in object type '%s'",
					fieldName, objectDef.Name),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		} else {
			// Check if the field type is primitive/scalar
			if !r.isPrimitiveOrScalarType(field.Type, schema) {
				fieldTypeName := r.getTypeName(field.Type)
				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("Field '%s' specified in @key directive must be a primitive or scalar type, but is of type '%s'",
						fieldName, fieldTypeName),
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

// extractFieldsString extracts the fields string from a GraphQL value
func (r *KeyDirectivesLint) extractFieldsString(value *ast.Value) string {
	if value == nil {
		return ""
	}

	switch value.Kind {
	case ast.StringValue:
		return value.Raw
	default:
		return ""
	}
}

// isPrimitiveOrScalarType checks if a type is a primitive or scalar type
func (r *KeyDirectivesLint) isPrimitiveOrScalarType(fieldType *ast.Type, schema *ast.Schema) bool {
	// If it's a list type, it's not allowed in @key
	if fieldType.Elem != nil {
		return false
	}

	// Get the underlying type name (remove NonNull wrapper)
	typeName := r.getTypeName(fieldType)

	// Check if it's a built-in scalar type
	if r.isBuiltInScalar(typeName) {
		return true
	}

	// Check if it's a custom scalar type defined in the schema
	for _, def := range schema.Types {
		if def.Name == typeName && def.Kind == ast.Scalar {
			return true
		}
	}

	return false
}

// getTypeName extracts the type name from a Type, removing List and NonNull wrappers
func (r *KeyDirectivesLint) getTypeName(fieldType *ast.Type) string {
	if fieldType == nil {
		return ""
	}

	// Handle NonNull wrapper
	if fieldType.NonNull {
		return r.getTypeName(&ast.Type{
			NamedType: fieldType.NamedType,
			Elem:      fieldType.Elem,
		})
	}

	// Handle List wrapper
	if fieldType.Elem != nil {
		return r.getTypeName(fieldType.Elem)
	}

	// Return the named type
	return fieldType.NamedType
}

// isBuiltInScalar checks if a type name is a built-in GraphQL scalar
func (r *KeyDirectivesLint) isBuiltInScalar(typeName string) bool {
	builtInScalars := map[string]bool{
		"String":  true,
		"Int":     true,
		"Float":   true,
		"Boolean": true,
		"ID":      true,
	}
	return builtInScalars[typeName]
}

// hasCommaSeparatedFields checks if the fields string contains commas indicating comma-separated fields
func (r *KeyDirectivesLint) hasCommaSeparatedFields(fieldsString string) bool {
	// Remove quotes if present
	trimmed := strings.Trim(fieldsString, `"`)

	// Check for commas that are not inside nested braces/brackets
	braceLevel := 0
	bracketLevel := 0

	for _, char := range trimmed {
		switch char {
		case '{':
			braceLevel++
		case '}':
			braceLevel--
		case '[':
			bracketLevel++
		case ']':
			bracketLevel--
		case ',':
			// If we find a comma at the top level (not inside nested structures), it's invalid
			if braceLevel == 0 && bracketLevel == 0 {
				return true
			}
		}
	}

	return false
}

// validateResolvableFalseConstraints checks that objects with resolvable: false @key directive have only one @key
func (r *KeyDirectivesLint) validateResolvableFalseConstraints(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all object types in the schema
	for _, def := range schema.Types {
		if def.Kind == ast.Object {
			errors = append(errors, r.checkResolvableFalseKeyDirectives(def, source)...)
		}
	}

	return errors
}

// checkResolvableFalseKeyDirectives checks if an object type has resolvable: false @key constraints
func (r *KeyDirectivesLint) checkResolvableFalseKeyDirectives(objectDef *ast.Definition, source *ast.Source) []types.LintError {
	var errors []types.LintError
	var keyDirectives []*ast.Directive
	var hasResolvableFalse bool

	// Find all @key directives on this object
	for _, directive := range objectDef.Directives {
		if directive.Name == "key" {
			keyDirectives = append(keyDirectives, directive)

			// Check if this @key directive has resolvable: false
			if r.hasResolvableFalse(directive) {
				hasResolvableFalse = true
			}
		}
	}

	// If any @key directive has resolvable: false, ensure there's only one @key directive total
	if hasResolvableFalse && len(keyDirectives) > 1 {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Object type '%s' has a @key directive with 'resolvable: false' but also has %d total @key directives. Objects with 'resolvable: false' @key directive can only have one @key directive.", objectDef.Name, len(keyDirectives)),
			Location: types.Location{
				Line:   objectDef.Position.Line,
				Column: objectDef.Position.Column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	return errors
}

// hasResolvableFalse checks if a @key directive has resolvable: false argument
func (r *KeyDirectivesLint) hasResolvableFalse(directive *ast.Directive) bool {
	// Look for the "resolvable" argument
	for _, arg := range directive.Arguments {
		if arg.Name == "resolvable" {
			// Check if the value is false
			if arg.Value != nil && arg.Value.Kind == ast.BooleanValue {
				return arg.Value.Raw == "false"
			}
		}
	}
	return false
}

// validateResolvableFalseIncludesAllFields checks that objects with single resolvable: false @key include all fields
func (r *KeyDirectivesLint) validateResolvableFalseIncludesAllFields(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check all object types in the schema
	for _, def := range schema.Types {
		if def.Kind == ast.Object {
			errors = append(errors, r.checkResolvableFalseIncludesAllFields(def, source, schema)...)
		}
	}

	return errors
}

// checkResolvableFalseIncludesAllFields validates that a single resolvable: false @key includes all object fields
func (r *KeyDirectivesLint) checkResolvableFalseIncludesAllFields(objectDef *ast.Definition, source *ast.Source, schema *ast.Schema) []types.LintError {
	var errors []types.LintError
	var keyDirectives []*ast.Directive

	// Find all @key directives on this object
	for _, directive := range objectDef.Directives {
		if directive.Name == "key" {
			keyDirectives = append(keyDirectives, directive)
		}
	}

	// Only check if there's exactly one @key directive with resolvable: false
	if len(keyDirectives) == 1 && r.hasResolvableFalse(keyDirectives[0]) {
		keyDirective := keyDirectives[0]

		// Get the fields argument from the @key directive
		var fieldsArg *ast.Argument
		for _, arg := range keyDirective.Arguments {
			if arg.Name == "fields" {
				fieldsArg = arg
				break
			}
		}

		if fieldsArg == nil {
			return errors // Skip if no fields argument (will be caught by other validation)
		}

		// Extract the fields string from the @key directive
		fieldsString := r.extractFieldsString(fieldsArg.Value)
		if fieldsString == "" {
			return errors // Skip if empty fields (will be caught by other validation)
		}

		// Parse the key fields to get individual field names using fragment parsing
		keyFields := r.parseResolvableFalseKeyFields(fieldsString, objectDef)

		// Get all field names from the object definition
		objectFieldNames := make(map[string]bool)
		for _, field := range objectDef.Fields {
			objectFieldNames[field.Name] = true
		}

		// Check if all object fields are included in the key fields
		missingFields := []string{}
		for fieldName := range objectFieldNames {
			if !contains(keyFields, fieldName) {
				missingFields = append(missingFields, fieldName)
			}
		}

		// Sort missing fields for consistent error messages
		sort.Strings(missingFields)

		// Report error if any fields are missing from the key
		if len(missingFields) > 0 {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Object type '%s' has a single @key directive with 'resolvable: false' but the key does not include all object fields. Missing fields in @key: %v. All fields must be included when using 'resolvable: false'.", objectDef.Name, missingFields),
				Location: types.Location{
					Line:   keyDirective.Position.Line,
					Column: keyDirective.Position.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}

// parseResolvableFalseKeyFields extracts individual field names from a fields string using fragment parsing
func (r *KeyDirectivesLint) parseResolvableFalseKeyFields(fieldsString string, objectDef *ast.Definition) []string {
	// Use the same fragment parsing approach as the main validation
	query := fmt.Sprintf("fragment x on %s { %s }", objectDef.Name, fieldsString)
	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		// If parsing fails, return empty slice (error will be caught by other validation)
		return []string{}
	}

	var result []string
	selectionSet := doc.Fragments[0].SelectionSet
	for _, sel := range selectionSet {
		fieldSel, ok := sel.(*ast.Field)
		if !ok {
			continue
		}
		// Only collect top-level field names (ignore nested selections for this validation)
		result = append(result, fieldSel.Name)
	}

	return result
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
