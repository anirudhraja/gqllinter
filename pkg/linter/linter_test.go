package linter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test schema content for various test scenarios
const (
	validSchema = `
		"""A user in the system"""
		type User {
			"""User's unique identifier"""
			id: ID!
			"""User's name"""
			name: String!
		}

		"""Query root type"""
		type Query {
			"""Get a user by ID"""
			user(id: ID!): User
		}
	`

	invalidSchema = `
		type User {
			id: ID!
			name: String!
		}

		type Query {
			getUser: User
		}
	`

	malformedSchema = `
		type User {
			id: ID!
			name: String!
		
		type Query {
			user: User
		}
	`
)

func TestNew(t *testing.T) {
	linter := New()

	if linter == nil {
		t.Fatal("Expected linter to be created, got nil")
	}

	if len(linter.rules) == 0 {
		t.Error("Expected linter to have built-in rules")
	}

	// Check that all expected rules are loaded
	expectedRuleCount := 36 // Based on the rules in the New() function
	if len(linter.rules) != expectedRuleCount {
		t.Errorf("Expected %d rules, got %d", expectedRuleCount, len(linter.rules))
	}

	// Verify enabledRules map is initialized
	if linter.enabledRules == nil {
		t.Error("Expected enabledRules map to be initialized")
	}
}

func TestGetAvailableRules(t *testing.T) {
	linter := New()
	rules := linter.GetAvailableRules()

	if len(rules) == 0 {
		t.Error("Expected available rules to be returned")
	}

	// Check for some expected rule names
	expectedRules := []string{
		"types-have-descriptions",
		"fields-have-descriptions",
		"no-hashtag-description",
		"naming-convention",
		"alphabetize",
	}

	ruleMap := make(map[string]bool)
	for _, rule := range rules {
		ruleMap[rule] = true
	}

	for _, expectedRule := range expectedRules {
		if !ruleMap[expectedRule] {
			t.Errorf("Expected rule %s not found in available rules", expectedRule)
		}
	}
}

func TestSetRules(t *testing.T) {
	linter := New()

	// Test setting specific rules
	rulesToEnable := []string{"types-have-descriptions", "fields-have-descriptions"}
	linter.SetRules(rulesToEnable)

	// Check that only specified rules are enabled
	for _, rule := range rulesToEnable {
		if !linter.enabledRules[rule] {
			t.Errorf("Expected rule %s to be enabled", rule)
		}
	}

	// Check that other rules are not enabled
	if linter.enabledRules["naming-convention"] {
		t.Error("Expected naming-convention rule to not be enabled")
	}

	// Test setting empty rules (should clear all)
	linter.SetRules([]string{})
	if len(linter.enabledRules) != 0 {
		t.Error("Expected all rules to be disabled when empty slice is provided")
	}
}

func TestParseSchemaFile(t *testing.T) {
	linter := New()

	t.Run("should parse valid schema file", func(t *testing.T) {
		// Create temporary file with valid schema
		tmpFile, err := createTempSchemaFile(t, validSchema)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		schema, source, err := linter.parseSchemaFile(tmpFile)
		if err != nil {
			t.Errorf("Expected no error parsing valid schema, got: %v", err)
		}

		if schema == nil {
			t.Error("Expected schema to be parsed, got nil")
		}

		if source == nil {
			t.Error("Expected source to be returned, got nil")
		} else if source.Name != tmpFile {
			t.Errorf("Expected source name to be %s, got %s", tmpFile, source.Name)
		}
	})

	t.Run("should fail on non-existent file", func(t *testing.T) {
		_, _, err := linter.parseSchemaFile("non-existent-file.graphql")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}

		if !strings.Contains(err.Error(), "failed to read file") {
			t.Errorf("Expected 'failed to read file' error, got: %v", err)
		}
	})

	t.Run("should fail on malformed schema", func(t *testing.T) {
		tmpFile, err := createTempSchemaFile(t, malformedSchema)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		_, _, err = linter.parseSchemaFile(tmpFile)
		if err == nil {
			t.Error("Expected error for malformed schema")
		}

		if !strings.Contains(err.Error(), "failed to parse schema") {
			t.Errorf("Expected 'failed to parse schema' error, got: %v", err)
		}
	})
}

func TestLintFile(t *testing.T) {
	linter := New()

	t.Run("should lint valid schema with minimal errors", func(t *testing.T) {
		tmpFile, err := createTempSchemaFile(t, validSchema)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		errors, err := linter.LintFile(tmpFile)
		if err != nil {
			t.Errorf("Expected no error linting file, got: %v", err)
		}

		// Valid schema should have minimal errors (mainly introspection field descriptions)
		if len(errors) == 0 {
			t.Log("No linting errors found in valid schema")
		}
	})

	t.Run("should lint invalid schema with multiple errors", func(t *testing.T) {
		tmpFile, err := createTempSchemaFile(t, invalidSchema)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		errors, err := linter.LintFile(tmpFile)
		if err != nil {
			t.Errorf("Expected no error linting file, got: %v", err)
		}

		if len(errors) == 0 {
			t.Error("Expected linting errors for invalid schema")
		}

		// Check for specific expected errors
		errorMessages := make([]string, len(errors))
		for i, e := range errors {
			errorMessages[i] = e.Message
		}

		// Should have errors for missing descriptions and query prefix
		hasDescriptionError := false
		hasPrefixError := false

		for _, msg := range errorMessages {
			if strings.Contains(msg, "missing a description") {
				hasDescriptionError = true
			}
			if strings.Contains(msg, "should not be prefixed with 'get'") {
				hasPrefixError = true
			}
		}

		if !hasDescriptionError {
			t.Error("Expected description error for invalid schema")
		}
		if !hasPrefixError {
			t.Error("Expected query prefix error for invalid schema")
		}
	})

	t.Run("should respect enabled rules filter", func(t *testing.T) {
		tmpFile, err := createTempSchemaFile(t, invalidSchema)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		// Enable only one rule
		linter.SetRules([]string{"types-have-descriptions"})

		errors, err := linter.LintFile(tmpFile)
		if err != nil {
			t.Errorf("Expected no error linting file, got: %v", err)
		}

		// Should only get errors from the enabled rule
		for _, e := range errors {
			if e.Rule != "types-have-descriptions" {
				t.Errorf("Expected only 'types-have-descriptions' errors, got error from rule: %s", e.Rule)
			}
		}
	})

	t.Run("should fail on non-existent file", func(t *testing.T) {
		_, err := linter.LintFile("non-existent-file.graphql")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("should fail on malformed schema", func(t *testing.T) {
		tmpFile, err := createTempSchemaFile(t, malformedSchema)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile) }()

		_, err = linter.LintFile(tmpFile)
		if err == nil {
			t.Error("Expected error for malformed schema")
		}
	})
}

func TestLoadCustomRules(t *testing.T) {
	linter := New()
	initialRuleCount := len(linter.rules)

	t.Run("should handle non-existent directory", func(t *testing.T) {
		err := linter.LoadCustomRules("non-existent-directory")
		if err != nil {
			t.Errorf("Expected no error for non-existent directory, got: %v", err)
		}

		// Rule count should remain the same
		if len(linter.rules) != initialRuleCount {
			t.Error("Expected rule count to remain the same for non-existent directory")
		}
	})

	t.Run("should handle empty directory", func(t *testing.T) {
		// Create temporary directory
		tmpDir, err := os.MkdirTemp("", "test-custom-rules")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		err = linter.LoadCustomRules(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for empty directory, got: %v", err)
		}

		// Rule count should remain the same
		if len(linter.rules) != initialRuleCount {
			t.Error("Expected rule count to remain the same for empty directory")
		}
	})

	t.Run("should handle directory with non-plugin files", func(t *testing.T) {
		// Create temporary directory with non-plugin files
		tmpDir, err := os.MkdirTemp("", "test-custom-rules")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		// Create a non-plugin file
		nonPluginFile := filepath.Join(tmpDir, "not-a-plugin.txt")
		err = os.WriteFile(nonPluginFile, []byte("not a plugin"), 0644)
		if err != nil {
			t.Fatalf("Failed to create non-plugin file: %v", err)
		}

		err = linter.LoadCustomRules(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for directory with non-plugin files, got: %v", err)
		}

		// Rule count should remain the same
		if len(linter.rules) != initialRuleCount {
			t.Error("Expected rule count to remain the same for directory with non-plugin files")
		}
	})
}

func TestLoadPlugin(t *testing.T) {
	linter := New()

	t.Run("should fail on non-existent plugin file", func(t *testing.T) {
		err := linter.loadPlugin("non-existent-plugin.so")
		if err == nil {
			t.Error("Expected error for non-existent plugin file")
		}

		if !strings.Contains(err.Error(), "failed to open plugin") {
			t.Errorf("Expected 'failed to open plugin' error, got: %v", err)
		}
	})

	t.Run("should fail on invalid plugin file", func(t *testing.T) {
		// Create a temporary file that's not a valid plugin
		tmpFile, err := os.CreateTemp("", "invalid-plugin-*.so")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		// Write some invalid content
		_, err = tmpFile.WriteString("not a valid plugin")
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		_ = tmpFile.Close()

		err = linter.loadPlugin(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for invalid plugin file")
		}
	})
}

// Helper function to create temporary schema files for testing
func createTempSchemaFile(t *testing.T, content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "test-schema-*.graphql")
	if err != nil {
		return "", err
	}

	_, err = tmpFile.WriteString(content)
	if err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", err
	}

	_ = tmpFile.Close()
	return tmpFile.Name(), nil
}

// Benchmark tests
func BenchmarkLintFile(b *testing.B) {
	linter := New()
	tmpFile, err := createTempSchemaFile(nil, validSchema)
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := linter.LintFile(tmpFile)
		if err != nil {
			b.Errorf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkParseSchemaFile(b *testing.B) {
	linter := New()
	tmpFile, err := createTempSchemaFile(nil, validSchema)
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := linter.parseSchemaFile(tmpFile)
		if err != nil {
			b.Errorf("Benchmark failed: %v", err)
		}
	}
}

// Integration test with actual rule execution
func TestLinterIntegration(t *testing.T) {
	linter := New()

	// Test schema with known violations
	testSchema := `
		type User {
			id: ID!
			userName: String!
		}

		type Query {
			getUser: User
		}
	`

	tmpFile, err := createTempSchemaFile(t, testSchema)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	errors, err := linter.LintFile(tmpFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify we get expected error types
	ruleErrorCounts := make(map[string]int)
	for _, e := range errors {
		ruleErrorCounts[e.Rule]++
	}

	// Should have errors from multiple rules
	expectedRules := []string{
		"types-have-descriptions",
		"fields-have-descriptions",
		"no-field-namespacing",
		"no-query-prefixes",
	}

	for _, expectedRule := range expectedRules {
		if ruleErrorCounts[expectedRule] == 0 {
			t.Errorf("Expected errors from rule %s, but got none", expectedRule)
		}
	}

	t.Logf("Integration test found %d total errors across %d different rules",
		len(errors), len(ruleErrorCounts))
}
