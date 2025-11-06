package rules

import "testing"

func TestQueryResponseNullable(t *testing.T) {
	rule := NewQueryResponseNullable()

	t.Run("should flag non-nullable top-level query fields (named and list types)", func(t *testing.T) {
		schema := `
		type Query {
			user: User!
			users: [User!]!
			ids: [ID!]!
		}

		type User {
			id: ID!
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "query-response-nullable") != 3 {
			t.Errorf("Expected exactly 3 errors for non-nullable top-level Query fields, got %d", countRuleErrors(errors, "query-response-nullable"))
		}

		expectedMessages := []string{
			"Query root field `user` should be nullable (`User` instead of `User!`) to prevent nulling out entire query response due to missing data.",
			"Query root field `users` should be nullable (`[User!]` instead of `[User!]!`) to prevent nulling out entire query response due to missing data.",
			"Query root field `ids` should be nullable (`[ID!]` instead of `[ID!]!`) to prevent nulling out entire query response due to missing data.",
		}

		for _, expectedMessage := range expectedMessages {
			if !containsError(errors, expectedMessage) {
				t.Errorf("Expected error message: %s", expectedMessage)
			}
		}
	})

	t.Run("should pass when all top-level query fields are nullable", func(t *testing.T) {
		schema := `
		type Query {
			user: User
			users: [User!]
			ids: [ID!]
			name: String
			count: Int
		}

		type User {
			id: ID!
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "query-response-nullable") > 0 {
			t.Errorf("Expected no errors for nullable top-level Query fields, got %d", countRuleErrors(errors, "query-response-nullable"))
		}
	})

	t.Run("should ignore inner field nullability and only enforce top-level", func(t *testing.T) {
		schema := `
		type Query {
			user: User
			users: [User!]
			matrix: [[User!]!]
		}

		type User {
			id: ID!
			name: String!
			email: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "query-response-nullable") != 0 {
			t.Errorf("Expected no errors when only inner fields are non-nullable, got %d", countRuleErrors(errors, "query-response-nullable"))
		}
	})

	t.Run("should flag a nested list that is non-nullable at the top level", func(t *testing.T) {
		schema := `
		type Query {
			grid: [[User!]!]!
		}

		type User {
			id: ID!
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "query-response-nullable") != 1 {
			t.Errorf("Expected exactly 1 error for non-nullable top-level nested list, got %d", countRuleErrors(errors, "query-response-nullable"))
		}

		expectedMessage := "Query root field `grid` should be nullable (`[[User!]!]` instead of `[[User!]!]!`) to prevent nulling out entire query response due to missing data."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})
}
