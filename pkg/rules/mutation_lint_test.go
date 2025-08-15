package rules

import (
	"strings"
	"testing"

	"github.com/nishant-rn/gqlparser/v2"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

func TestMutationLint(t *testing.T) {
	rule := NewMutationLint()

	tests := []struct {
		name           string
		schema         string
		expectedErrors int
		expectedMsg    string
	}{
		{
			name: "Valid: Complete mutation response union pattern",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union MobileRiders @responseUnion = MockMobileRider | RiderNotFound

				type RiderNotFound @error {
					code: String!
					message: String!
				}

				type MockMobileRider @key(fields: "id") {
					id: ID!
					user: MockUser!
					mobile_number: String!
					name: String!
				}

				type MockUser {
					id: ID!
					name: String!
				}

				type Mutation {
					resolveMobileRiders(id: ID!): MobileRiders
				}

				directive @key(fields: String!) on OBJECT
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Mutation returns non-union type",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				type MockMobileRider {
					id: ID!
					name: String!
				}

				type Mutation {
					resolveMobileRiders(id: ID!): MockMobileRider
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Mutation field 'resolveMobileRiders' must return a union type with @responseUnion directive, but returns 'MockMobileRider'",
		},
		{
			name: "Invalid: Mutation returns union without @responseUnion directive",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union MobileRiders = MockMobileRider | RiderNotFound

				type RiderNotFound @error {
					code: String!
					message: String!
				}

				type MockMobileRider {
					id: ID!
					name: String!
				}

				type Mutation {
					resolveMobileRiders(id: ID!): MobileRiders
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Mutation field 'resolveMobileRiders' returns union 'MobileRiders' which must have @responseUnion directive",
		},
		{
			name: "Invalid: @responseUnion union has no success types",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union MobileRiders @responseUnion = RiderNotFound | AnotherError

				type RiderNotFound @error {
					code: String!
					message: String!
				}

				type AnotherError @error {
					code: String!
					message: String!
				}

				type Mutation {
					resolveMobileRiders(id: ID!): MobileRiders
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Union 'MobileRiders' with @responseUnion directive must have exactly one success type (non-@error type), but has none",
		},
		{
			name: "Invalid: @responseUnion union has multiple success types",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union MobileRiders @responseUnion = MockMobileRider | AnotherSuccess | RiderNotFound

				type RiderNotFound @error {
					code: String!
					message: String!
				}

				type MockMobileRider {
					id: ID!
					name: String!
				}

				type AnotherSuccess {
					id: ID!
					data: String!
				}

				type Mutation {
					resolveMobileRiders(id: ID!): MobileRiders
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Union 'MobileRiders' with @responseUnion directive must have exactly one success type (non-@error type), but has 2",
		},
		{
			name: "Valid: Multiple mutations with proper @responseUnion unions",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union MobileRiders @responseUnion = MockMobileRider | RiderNotFound
				union UserResult @responseUnion = User | UserNotFound

				type RiderNotFound @error {
					code: String!
					message: String!
				}

				type UserNotFound @error {
					code: String!
					message: String!
				}

				type MockMobileRider {
					id: ID!
					name: String!
				}

				type User {
					id: ID!
					email: String!
				}

				type Mutation {
					resolveMobileRiders(id: ID!): MobileRiders
					getUser(id: ID!): UserResult
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: @error types used in query responses",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union UserResult @responseUnion = User | UserNotFound

				type UserNotFound @error {
					code: String!
					message: String!
				}

				type User {
					id: ID!
					email: String!
				}

				type Query {
					getUser(id: ID!): UserResult
				}

				type Mutation {
					updateUser(id: ID!): UserResult
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: @error type used in union not returned by mutation or query",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union UnusedUnion = User | UserNotFound

				type UserNotFound @error {
					code: String!
					message: String!
				}

				type User {
					id: ID!
					email: String!
				}

				type SomeOtherType {
					field: UnusedUnion
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'UserNotFound' has @error directive but is used in union 'UnusedUnion' which is not returned by any mutation or query",
		},
		{
			name: "Valid: No mutations defined",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				type User {
					id: ID!
					email: String!
				}

				type Query {
					getUser(id: ID!): User
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Multiple validation errors in same schema",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union BadUnion1 @responseUnion = Error1 | Error2
				union BadUnion2 = Success1 | Error3

				type Error1 @error {
					code: String!
				}

				type Error2 @error {
					message: String!
				}

				type Error3 @error {
					code: String!
				}

				type Success1 {
					id: ID!
				}

				type Mutation {
					mutation1(id: ID!): BadUnion1
					mutation2(id: ID!): BadUnion2
					mutation3(id: ID!): Success1
				}
			`,
			expectedErrors: 3,
		},
		{
			name: "Invalid: @responseUnion union with valid success count but non-@error other types",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT

				union TestUnion @responseUnion = SuccessType | NonErrorButNotSuccess

				type SuccessType {
					id: ID!
					data: String!
				}

				type NonErrorButNotSuccess {
					id: ID!
					other: String!
				}

				type Mutation {
					testMutation(id: ID!): TestUnion
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Union 'TestUnion' with @responseUnion directive must have exactly one success type (non-@error type), but has 2",
		},
		{
			name: "Valid: Complex valid schema with multiple patterns",
			schema: `
				directive @responseUnion on UNION
				directive @error on OBJECT
				directive @key(fields: String!) on OBJECT

				union CreateUserResult @responseUnion = User | ValidationError | DatabaseError
				union UpdateUserResult @responseUnion = User | UserNotFound | ValidationError

				type ValidationError @error {
					code: String!
					message: String!
					field: String
				}

				type DatabaseError @error {
					code: String!
					message: String!
					details: String
				}

				type UserNotFound @error {
					code: String!
					message: String!
					userId: ID!
				}

				type User @key(fields: "id") {
					id: ID!
					email: String!
					name: String!
					createdAt: String!
				}

				type Query {
					getUser(id: ID!): UpdateUserResult
				}

				type Mutation {
					createUser(email: String!, name: String!): CreateUserResult
					updateUser(id: ID!, name: String): UpdateUserResult
				}
			`,
			expectedErrors: 0,
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
			schema, gqlErr := gqlparser.LoadSchema(source)
			if gqlErr != nil {
				t.Fatalf("Failed to parse schema: %v", gqlErr)
			}

			// Run the rule
			errors := rule.Check(schema, source)

			// Check error count
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
					if strings.Contains(err.Message, tt.expectedMsg) {
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

			// Verify all errors have correct rule name
			for _, err := range errors {
				if err.Rule != rule.Name() {
					t.Errorf("Expected rule name '%s', got '%s'", rule.Name(), err.Rule)
				}
			}
		})
	}
}

func TestMutationLint_RuleMetadata(t *testing.T) {
	rule := NewMutationLint()

	// Test rule name
	expectedName := "mutation-lint"
	if rule.Name() != expectedName {
		t.Errorf("Expected rule name '%s', got '%s'", expectedName, rule.Name())
	}

	// Test rule description
	description := rule.Description()
	expectedDescKeywords := []string{"mutations", "responseUnion", "error", "unions", "success type"}
	for _, keyword := range expectedDescKeywords {
		if !strings.Contains(description, keyword) {
			t.Errorf("Expected description to contain '%s', got: %s", keyword, description)
		}
	}
}
