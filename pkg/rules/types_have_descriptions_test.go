package rules

import "testing"

func TestTypesHaveDescriptions(t *testing.T) {
	rule := NewTypesHaveDescriptions()

	t.Run("should flag types without descriptions", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if len(errors) == 0 {
			t.Error("Expected error for type without description")
		}
		if !containsError(errors, "The object type `User` is missing a description.") {
			t.Error("Expected specific error message about missing description")
		}
	})

	t.Run("should pass types with descriptions", func(t *testing.T) {
		schema := `
		"""A user in the system"""
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "types-have-descriptions") > 0 {
			t.Error("Expected no errors for type with description")
		}
	})

	t.Run("should pass builtin types without descriptions", func(t *testing.T) {
		schema := `
		type Query {
			id: ID!
			name: String
			active: Boolean
			age: Int
			score: Float
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "types-have-descriptions") > 0 {
			t.Error("Expected no errors for type with description")
		}
	})

	t.Run("should pass default root operation types without descriptions", func(t *testing.T) {
		schema := `
		type Query {
			hello: String
		}

		type Mutation {
			doThing: String
		}

		type Subscription {
			ping: String
		}

		"""A user in the system"""
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "types-have-descriptions") > 0 {
			t.Error("Expected no errors for type with description")
		}
	})

	t.Run("should pass custom root operation types without descriptions", func(t *testing.T) {
		schema := `
		schema {
			query: RootQuery
			mutation: RootMutation
			subscription: RootSubscription
		}

		type RootQuery {
			hello: String
		}

		type RootMutation {
			doThing: String
		}

		type RootSubscription {
			ping: String
		}

		"""A user in the system"""
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "types-have-descriptions") > 0 {
			t.Error("Expected no errors for type with description")
		}
	})
}
