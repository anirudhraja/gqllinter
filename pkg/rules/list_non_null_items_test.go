package rules

import (
	"testing"
)

func TestListNonNullItems(t *testing.T) {
	rule := NewListNonNullItems()

	t.Run("should flag lists with nullable items", func(t *testing.T) {
		schema := `
		type User {
			tags: [String]
			friends: [User]
			test1: [[String]!]
			test2: [[[String]!]]
			test3: [[[String!]!]!]
		}
		`
		errors := runRule(t, rule, schema)
		count := countRuleErrors(errors, "list-non-null-items")

		// Debug: print all errors
		t.Logf("Found %d errors:", count)
		for _, err := range errors {
			if err.Rule == "list-non-null-items" {
				t.Logf("  - %s", err.Message)
			}
		}

		if count != 4 {
			t.Errorf("Expected 4 errors for nullable list items, got %d", count)
		}
	})

	t.Run("should pass lists with non-null items", func(t *testing.T) {
		schema := `
		type User {
			tags: [String!]!
			friends: [User!]!
			users : [[User!]!]
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "list-non-null-items") > 0 {
			t.Error("Expected no list errors for non-null items")
		}
	})
}
