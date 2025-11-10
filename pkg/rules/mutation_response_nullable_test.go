package rules

import (
	"fmt"
	"strings"
	"testing"
)

func TestMutationResponseNullable(t *testing.T) {
	rule := NewMutationResponseNullable()

	t.Run("should flag root level non-null mutation response fields", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser: CreateUserResult!
		}

		type CreateUserResult {
			user: User!
			success: Boolean!
			message: String!
		}

		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") != 1 {
			t.Error("Expected exactly 1 error for non-null response fields")
		}

		if strings.Contains(errors[0].Message, "Mutation root field `createUser`") == false {
			t.Errorf("Expected error for non-null root response field createUser")
		}
	})

	t.Run("should pass valid mutation with nullable root fields irrespective of response type fields", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser: CreateUserResult
			updateUser: UpdateUserResult
		}

		type CreateUserResult {
			user: User
			success: Boolean!
			errors: [String]
		}

		type UpdateUserResult {
			user: User!
			message: String
		}

		type User {
			id: ID!
			name: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") > 0 {
			t.Error("Expected no errors for valid mutation structure")
		}
	})

	t.Run("should handle mutations returning scalars", func(t *testing.T) {
		schema := `
		type Mutation {
			deleteUser: Boolean
			getUserCount: Int
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") > 0 {
			t.Error("Expected no errors for scalar mutation returns")
		}
	})

	t.Run("should handle list types in mutation responses", func(t *testing.T) {
		schema := `
		type Mutation {
			createUsers: [CreateUsersResult!]!
			updateUsers: [CreateUsersResult!]
  			createProfile: [[String!]!]!
		}
	
		type CreateUsersResult {
			users: [User!]!
			errors: [String!]!
			successCount: Int!
		}
	
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") != 2 {
			t.Error("Expected exactly 2 errors for non-null list type root fields")
		}

		// Check specific error messages for response fields
		expectedFields := []string{"createUsers", "createProfile"}
		for _, field := range expectedFields {
			found := false
			for _, err := range errors {
				if err.Rule == "mutation-response-nullable" &&
					strings.Contains(err.Message, fmt.Sprintf("Mutation root field `%s`", field)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error for non-null root response field %s", field)
			}
		}
	})

	t.Run("should handle schema without mutations", func(t *testing.T) {
		schema := `
		type Query {
			user: User
		}

		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") > 0 {
			t.Error("Expected no errors for schema without mutations")
		}
	})

	t.Run("should handle union and interface return types", func(t *testing.T) {
		schema := `
		type Mutation {
			createContent: Content!
		}

		union Content = Article | Video

		type Article {
			id: ID!
			title: String!
		}

		type Video {
			id: ID!
			duration: Int!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") != 1 {
			t.Error("Expected at least 1 error for non-null root field of union type")
		}
	})
}
