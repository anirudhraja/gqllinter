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
	return "Types defined in a schema file should not be extended in the same file. Extensions should be in separate files."
}

// Check validates that types are not extended in the same file where they are defined
func (r *NoSameFileExtend) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Parse the source to find type definitions and extensions
	definedTypes := r.findDefinedTypes(source)
	extendedTypes := r.findExtendedTypes(source)

	// Check for conflicts
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

	return errors
}

// TypeInfo holds information about type definitions and extensions
type TypeInfo struct {
	Line   int
	Column int
	Name   string
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
			if typeName != "" {
				definedTypes[typeName] = TypeInfo{
					Line:   lineNum + 1, // Convert to 1-indexed
					Column: strings.Index(line, "type") + 1,
					Name:   typeName,
				}
			}
		}
	}

	return definedTypes
}

// findExtendedTypes parses the source to find all type extensions
func (r *NoSameFileExtend) findExtendedTypes(source *ast.Source) map[string]TypeInfo {
	extendedTypes := make(map[string]TypeInfo)
	lines := strings.Split(source.Input, "\n")

	for lineNum, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for type extensions
		if r.isTypeExtension(trimmedLine) {
			typeName := r.extractExtendTypeName(trimmedLine)
			if typeName != "" {
				extendedTypes[typeName] = TypeInfo{
					Line:   lineNum + 1, // Convert to 1-indexed
					Column: strings.Index(line, "extend") + 1,
					Name:   typeName,
				}
			}
		}
	}

	return extendedTypes
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

// isTypeExtension checks if a line contains a type extension
func (r *NoSameFileExtend) isTypeExtension(line string) bool {
	return strings.HasPrefix(line, "extend type ") ||
		strings.HasPrefix(line, "extend interface ") ||
		strings.HasPrefix(line, "extend input ") ||
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
