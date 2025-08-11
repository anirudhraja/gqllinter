package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// NoSameFileExtend checks that types defined in a file are not extended in the same file
type NoSameFileExtend struct{}

// NewNoSameFileExtend creates a new instance of the NoSameFileExtend rule
func NewNoSameFileExtend() *NoSameFileExtend {
	return &NoSameFileExtend{}
}

// Name returns the rule name
func (r *NoSameFileExtend) Name() string {
	return "no-same-file-extend"
}

// Description returns what this rule checks
func (r *NoSameFileExtend) Description() string {
	return "Types defined in a schema file should not be extended in the same file. Extensions should be in separate files. Additionally, only object types and interfaces can be extended. Extended object types must have the @key directive."
}

// Check validates that types are not extended in the same file where they are defined
// and that only object types and interfaces can be extended
func (r *NoSameFileExtend) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Parse the source to find type definitions and extensions
	definedTypes := r.findDefinedTypes(source)
	extendedTypes := r.findExtendedTypes(source)
	invalidExtensions := r.findInvalidExtensions(source)

	// Check for invalid extension types (only object and interface are allowed)
	for _, invalidExt := range invalidExtensions {
		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Cannot extend %s '%s' at line %d. Only object types and interfaces can be extended.",
				invalidExt.TypeKind, invalidExt.Name, invalidExt.Line),
			Location: types.Location{
				Line:   invalidExt.Line,
				Column: invalidExt.Column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	// Check for same-file extension conflicts
	for typeName, extendInfo := range extendedTypes {
		if defInfo, exists := definedTypes[typeName]; exists {
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Type '%s' is defined at line %d and extended at line %d in the same file. Types should not be extended in the same file where they are defined.",
					typeName, defInfo.Line, extendInfo.Line),
				Location: types.Location{
					Line:   extendInfo.Line,
					Column: extendInfo.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	// Check that extended object types have @key directive
	keyDirectiveErrors := r.checkExtendedTypesHaveKeyDirective(schema, source, extendedTypes)
	errors = append(errors, keyDirectiveErrors...)

	return errors
}

// TypeInfo holds information about type definitions and extensions
type TypeInfo struct {
	Line     int
	Column   int
	Name     string
	TypeKind string
}

// findDefinedTypes parses the source to find all type definitions
func (r *NoSameFileExtend) findDefinedTypes(source *ast.Source) map[string]TypeInfo {
	definedTypes := make(map[string]TypeInfo)
	lines := strings.Split(source.Input, "\n")

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for type definitions (but not extensions)
		if r.isTypeDefinition(trimmedLine) && !r.isTypeExtension(trimmedLine) {
			typeName := r.extractTypeName(trimmedLine)
			typeKind := r.extractTypeKind(trimmedLine)
			if typeName != "" {
				keywordIndex := strings.Index(line, typeKind)
				if keywordIndex == -1 {
					keywordIndex = 0
				}
				definedTypes[typeName] = TypeInfo{
					Line:     lineNum + 1, // Convert to 1-indexed
					Column:   keywordIndex + 1,
					Name:     typeName,
					TypeKind: typeKind,
				}
			}
		}
	}

	return definedTypes
}

// findExtendedTypes parses the source to find all valid type extensions (only type and interface)
func (r *NoSameFileExtend) findExtendedTypes(source *ast.Source) map[string]TypeInfo {
	extendedTypes := make(map[string]TypeInfo)
	lines := strings.Split(source.Input, "\n")

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for valid type extensions (only type and interface)
		if r.isValidTypeExtension(trimmedLine) {
			typeName := r.extractExtendTypeName(trimmedLine)
			typeKind := r.extractExtendTypeKind(trimmedLine)
			if typeName != "" {
				extendedTypes[typeName] = TypeInfo{
					Line:     lineNum + 1, // Convert to 1-indexed
					Column:   strings.Index(line, "extend") + 1,
					Name:     typeName,
					TypeKind: typeKind,
				}
			}
		}
	}

	return extendedTypes
}

// findInvalidExtensions finds extensions of types other than object and interface
func (r *NoSameFileExtend) findInvalidExtensions(source *ast.Source) []TypeInfo {
	var invalidExtensions []TypeInfo
	lines := strings.Split(source.Input, "\n")

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for invalid type extensions (input, enum, union, scalar)
		if r.isInvalidTypeExtension(trimmedLine) {
			typeName := r.extractExtendTypeName(trimmedLine)
			typeKind := r.extractExtendTypeKind(trimmedLine)
			if typeName != "" {
				invalidExtensions = append(invalidExtensions, TypeInfo{
					Line:     lineNum + 1, // Convert to 1-indexed
					Column:   strings.Index(line, "extend") + 1,
					Name:     typeName,
					TypeKind: typeKind,
				})
			}
		}
	}

	return invalidExtensions
}

// isTypeDefinition checks if a line contains a type definition
func (r *NoSameFileExtend) isTypeDefinition(line string) bool {
	return strings.HasPrefix(line, "type ") ||
		strings.HasPrefix(line, "interface ") ||
		strings.HasPrefix(line, "input ") ||
		strings.HasPrefix(line, "enum ") ||
		strings.HasPrefix(line, "union ") ||
		strings.HasPrefix(line, "scalar ")
}

// isTypeExtension checks if a line contains any type extension
func (r *NoSameFileExtend) isTypeExtension(line string) bool {
	return strings.HasPrefix(line, "extend type ") ||
		strings.HasPrefix(line, "extend interface ") ||
		strings.HasPrefix(line, "extend input ") ||
		strings.HasPrefix(line, "extend enum ") ||
		strings.HasPrefix(line, "extend union ") ||
		strings.HasPrefix(line, "extend scalar ")
}

// isValidTypeExtension checks if a line contains a valid type extension (only type and interface)
func (r *NoSameFileExtend) isValidTypeExtension(line string) bool {
	return strings.HasPrefix(line, "extend type ") ||
		strings.HasPrefix(line, "extend interface ")
}

// isInvalidTypeExtension checks if a line contains an invalid type extension
func (r *NoSameFileExtend) isInvalidTypeExtension(line string) bool {
	return strings.HasPrefix(line, "extend input ") ||
		strings.HasPrefix(line, "extend enum ") ||
		strings.HasPrefix(line, "extend union ") ||
		strings.HasPrefix(line, "extend scalar ")
}

// extractTypeName extracts the type name from a type definition line
func (r *NoSameFileExtend) extractTypeName(line string) string {
	// Remove comments and extra whitespace
	if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
		line = line[:commentIndex]
	}
	line = strings.TrimSpace(line)

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return ""
	}

	// Get the type name (second part after the keyword)
	typeName := parts[1]

	// Remove any trailing characters like { or implements
	if bracketIndex := strings.Index(typeName, "{"); bracketIndex != -1 {
		typeName = typeName[:bracketIndex]
	}
	if implementsIndex := strings.Index(typeName, "implements"); implementsIndex != -1 {
		typeName = typeName[:implementsIndex]
	}

	return strings.TrimSpace(typeName)
}

// extractExtendTypeName extracts the type name from a type extension line
func (r *NoSameFileExtend) extractExtendTypeName(line string) string {
	// Remove comments and extra whitespace
	if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
		line = line[:commentIndex]
	}
	line = strings.TrimSpace(line)

	parts := strings.Fields(line)
	if len(parts) < 3 {
		return ""
	}

	// Get the type name (third part after "extend" and the type keyword)
	typeName := parts[2]

	// Remove any trailing characters like {
	if bracketIndex := strings.Index(typeName, "{"); bracketIndex != -1 {
		typeName = typeName[:bracketIndex]
	}

	return strings.TrimSpace(typeName)
}

// extractTypeKind extracts the type kind from a type definition line
func (r *NoSameFileExtend) extractTypeKind(line string) string {
	// Remove comments and extra whitespace
	if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
		line = line[:commentIndex]
	}
	line = strings.TrimSpace(line)

	parts := strings.Fields(line)
	if len(parts) < 1 {
		return ""
	}

	return parts[0] // First part is the type kind (type, interface, input, enum, union, scalar)
}

// extractExtendTypeKind extracts the type kind from a type extension line
func (r *NoSameFileExtend) extractExtendTypeKind(line string) string {
	// Remove comments and extra whitespace
	if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
		line = line[:commentIndex]
	}
	line = strings.TrimSpace(line)

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return ""
	}

	return parts[1] // Second part after "extend" is the type kind
}

// checkExtendedTypesHaveKeyDirective checks that extended object types have @key directive
func (r *NoSameFileExtend) checkExtendedTypesHaveKeyDirective(schema *ast.Schema, source *ast.Source, extendedTypes map[string]TypeInfo) []types.LintError {
	var errors []types.LintError

	for typeName, extendInfo := range extendedTypes {
		// Only check object types and interfaces
		if !(strings.ToLower(extendInfo.TypeKind) == "type" || strings.ToLower(extendInfo.TypeKind) == "interface") {
			continue
		}

		// Find the type in the schema to check for @key directive
		var targetType *ast.Definition
		for _, def := range schema.Types {
			if def.Name == typeName && (def.Kind == ast.Object || def.Kind == ast.Interface) {
				targetType = def
				break
			}
		}

		// If we can't find the type in the schema, skip (it might be defined elsewhere)
		if targetType == nil {
			continue
		}

		// Check if the type has @key directive
		hasKeyDirective := false
		for _, directive := range targetType.Directives {
			if directive.Name == "key" {
				hasKeyDirective = true
				break
			}
		}

		// If no @key directive found, report error
		if !hasKeyDirective {
			typeKind := "object"
			if targetType.Kind == ast.Interface {
				typeKind = "interface"
			}
			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("Extended %s type '%s' at line %d must have the @key directive.",
					typeKind, typeName, extendInfo.Line),
				Location: types.Location{
					Line:   extendInfo.Line,
					Column: extendInfo.Column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}
