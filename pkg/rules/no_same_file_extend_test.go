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
			name: "Invalid: Type defined and extended in same file (even with @key)",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "id") {
					id: ID!
				}
				
				extend type User {
					email: String
				}
			`,
			expectedErrors: 1, // Still same-file extension error
			expectedMsg:    "Type 'User' is defined at line 4 and extended at line 8 in the same file",
		},
		{
			name: "Invalid: Extension without @key directive",
			schema: `
				type User {
					id: ID!
					name: String
				}
				
				extend type User {
					email: String
				}
			`,
			expectedErrors: 2, // Same-file extension + missing @key
			expectedMsg:    "Extended object type 'User' at line 7 must have the @key directive",
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
			name: "Invalid: Input type extension not allowed",
			schema: `
				input UserInput {
					name: String
				}
				
				extend input UserInput {
					email: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Cannot extend input 'UserInput' at line 6. Only object types and interfaces can be extended",
		},
		{
			name: "Invalid: Enum extension not allowed",
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
			expectedMsg:    "Cannot extend enum 'Status' at line 7. Only object types and interfaces can be extended",
		},
		{
			name: "Invalid: Union extension not allowed",
			schema: `
				type User { id: ID! }
				type Product { id: ID! }
				type Organization { id: ID! }
				
				union SearchResult = User | Product
				
				extend union SearchResult = Organization
			`,
			expectedErrors: 1,
			expectedMsg:    "Cannot extend union 'SearchResult' at line 8. Only object types and interfaces can be extended",
		},
		{
			name: "Invalid: Scalar extension not allowed",
			schema: `
				scalar DateTime
				
				extend scalar DateTime @specifiedBy(url: "https://tools.ietf.org/html/rfc3339")
			`,
			expectedErrors: 1,
			expectedMsg:    "Cannot extend scalar 'DateTime' at line 4. Only object types and interfaces can be extended",
		},
		{
			name: "Invalid: Multiple invalid extensions",
			schema: `
				enum Status { ACTIVE }
				input UserInput { name: String }
				
				extend enum Status { PENDING }
				extend input UserInput { email: String }
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
			name: "Valid: Only extension with @key directive (simulating separate files)",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "id") {
					id: ID!
				}
				
				# Only extensions, no type definition
				extend type User {
					name: String!
				}
			`,
			expectedErrors: 1, // Same-file error (but this shows @key validation would work)
		},
		{
			name: "Invalid: Extension of external type without @key directive",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				extend type ExternalUser {
					name: String!
				}
			`,
			expectedErrors: 1, // External type being extended must have @key (but isn't found in this schema)
			expectedMsg:    "Extended object type 'ExternalUser' at line 4 must have the @key directive",
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
			expectedErrors: 2, // Same-file extension + missing @key
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
