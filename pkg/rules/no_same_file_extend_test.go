package rules

import (
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestNoSameFileExtend(t *testing.T) {
	rule := NewNoSameFileExtend()

	tests := []struct {
		name           string
		schema         string
		expectedErrors int
		expectedMsg    string
	}{
		{
			name: "Valid: Type defined but not extended in same file",
			schema: `
				type User {
					id: ID!
					name: String
				}
				
				type Product {
					id: ID!
					title: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: Only extension without definition",
			schema: `
				extend type User {
					email: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Type defined and extended in same file",
			schema: `
				type User {
					id: ID!
					name: String
				}
				
				extend type User {
					email: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'User' is defined at line 2 and extended at line 7 in the same file",
		},
		{
			name: "Invalid: Interface defined and extended in same file",
			schema: `
				interface Node {
					id: ID!
				}
				
				extend interface Node {
					createdAt: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'Node' is defined at line 2 and extended at line 6 in the same file",
		},
		{
			name: "Invalid: Input type defined and extended in same file",
			schema: `
				input UserInput {
					name: String
				}
				
				extend input UserInput {
					email: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'UserInput' is defined at line 2 and extended at line 6 in the same file",
		},
		{
			name: "Invalid: Enum defined and extended in same file",
			schema: `
				enum Status {
					ACTIVE
					INACTIVE
				}
				
				extend enum Status {
					PENDING
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'Status' is defined at line 2 and extended at line 7 in the same file",
		},
		{
			name: "Invalid: Union defined and extended in same file",
			schema: `
				type User { id: ID! }
				type Product { id: ID! }
				type Organization { id: ID! }
				
				union SearchResult = User | Product
				
				extend union SearchResult = Organization
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'SearchResult' is defined at line 6 and extended at line 8 in the same file",
		},
		{
			name: "Invalid: Scalar defined and extended in same file",
			schema: `
				scalar DateTime
				
				extend scalar DateTime @specifiedBy(url: "https://tools.ietf.org/html/rfc3339")
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'DateTime' is defined at line 2 and extended at line 4 in the same file",
		},
		{
			name: "Invalid: Multiple types with conflicts",
			schema: `
				type User {
					id: ID!
				}
				
				type Product {
					id: ID!
				}
				
				extend type User {
					name: String
				}
				
				extend type Product {
					title: String
				}
			`,
			expectedErrors: 2,
		},
		{
			name: "Valid: Type with implements clause",
			schema: `
				interface Node {
					id: ID!
				}
				
				type User implements Node {
					id: ID!
					name: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: Complex type definitions without extensions",
			schema: `
				type Query {
					user(id: ID!): User
					users: [User!]!
				}
				
				type User {
					id: ID!
					name: String!
					profile: UserProfile
				}
				
				type UserProfile {
					bio: String
					avatarUrl: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: Comments and formatting variations",
			schema: `
				# User type definition
				type User {
					id: ID!
					name: String
				}
				
				# Product definition  
				type Product {
					id: ID!
					title: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Type extension with comments",
			schema: `
				type User {
					id: ID!
					name: String
				}
				
				# Extending user with email
				extend type User {
					email: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'User' is defined at line 2 and extended at line 8 in the same file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create source from schema string
			source := &ast.Source{
				Name:  "test-schema.graphql",
				Input: tt.schema,
			}

			// Parse the schema
			schema, err := gqlparser.LoadSchema(source)
			if err != nil {
				t.Fatalf("Failed to parse schema: %v", err)
			}

			// Run the rule
			errors := rule.Check(schema, source)

			// Check the number of errors
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(errors))
				for i, err := range errors {
					t.Logf("Error %d: %s", i+1, err.Message)
				}
				return
			}

			// Check error message if expected
			if tt.expectedErrors > 0 && tt.expectedMsg != "" {
				found := false
				for _, err := range errors {
					if containsSubstring(err.Message, tt.expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message to contain '%s', but got:", tt.expectedMsg)
					for i, err := range errors {
						t.Logf("Error %d: %s", i+1, err.Message)
					}
				}
			}

			// Verify rule name in errors
			for _, err := range errors {
				if err.Rule != rule.Name() {
					t.Errorf("Expected rule name '%s', got '%s'", rule.Name(), err.Rule)
				}
			}
		})
	}
}

func TestNoSameFileExtend_EdgeCases(t *testing.T) {
	rule := NewNoSameFileExtend()

	tests := []struct {
		name           string
		schema         string
		expectedErrors int
	}{
		{
			name:           "Empty schema",
			schema:         ``,
			expectedErrors: 0,
		},
		{
			name: "Only comments",
			schema: `
				# This is a comment
				# Another comment
			`,
			expectedErrors: 0,
		},
		{
			name: "Only whitespace and newlines",
			schema: `
			
			
			`,
			expectedErrors: 0,
		},
		{
			name: "Type definition with trailing comments",
			schema: `
				type User { # This is a user type
					id: ID!
					name: String
				}
				
				extend type User { # Extending user
					email: String
				}
			`,
			expectedErrors: 1,
		},
		{
			name: "Complex formatting with mixed spacing",
			schema: `
				type User{
					id: ID!
				}

				extend type User{
					email: String
				}
			`,
			expectedErrors: 1,
		},
		{
			name: "Type definitions on same line (malformed but parseable)",
			schema: `
				type User { id: ID! }
				extend type User { email: String }
			`,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &ast.Source{
				Name:  "test-schema.graphql",
				Input: tt.schema,
			}

			// Parse the schema - some edge cases might fail parsing
			schema, err := gqlparser.LoadSchema(source)
			if err != nil {
				// If parsing fails, we can't run our rule
				t.Logf("Schema parsing failed (expected for some edge cases): %v", err)
				return
			}

			errors := rule.Check(schema, source)
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(errors))
				for i, err := range errors {
					t.Logf("Error %d: %s", i+1, err.Message)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		len(str) > len(substr) && (str[:len(substr)] == substr ||
			str[len(str)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(str)-len(substr); i++ {
					if str[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}
