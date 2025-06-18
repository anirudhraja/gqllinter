# GraphQL Linter Rules Tests

This directory contains comprehensive tests for all GraphQL linting rules implemented in this project.

## Running Tests

### Run all rule tests
```bash
go test ./pkg/rules -v
```

### Run tests from the rules directory
```bash
cd pkg/rules
go test -v
```

### Run tests for the entire project
```bash
go test ./...
```

## Test Structure

### Test Files
- `rules_test.go` - Comprehensive test suite for all rules
- `testdata/test_schema.graphql` - Test schema with intentional violations

### Test Coverage

The test suite covers all 20 implemented rules:

#### Documentation Rules (5)
- `TestTypesHaveDescriptions` - Ensures all types have descriptions
- `TestFieldsHaveDescriptions` - Ensures all fields have descriptions  
- `TestNoHashtagDescription` - Prevents hashtag comments
- `TestCapitalizedDescriptions` - Ensures descriptions start with capital letters
- `TestEnumDescriptions` - Ensures enum values have descriptions

#### Naming Rules (5)
- `TestNamingConvention` - Validates naming conventions
- `TestNoFieldNamespacing` - Prevents redundant field prefixes
- `TestNoQueryPrefixes` - Prevents get/list/find prefixes in queries
- `TestInputName` - Validates mutation input naming
- `TestInputEnumSuffix` - Ensures input enums have Input suffix

#### Schema Design Rules (4)
- `TestMinimalTopLevelQueries` - Limits top-level query fields
- `TestNoUnusedFields` - Detects unused fields
- `TestNoUnusedTypes` - Detects unused types
- `TestEnumUnknownCase` - Requires UNKNOWN case in output enums

#### Schema Evolution Rules (3)
- `TestRequireDeprecationReason` - Requires deprecation reasons
- `TestNoScalarResultTypeOnMutation` - Prevents scalar mutation returns
- `TestMutationResponseNullable` - Ensures mutation responses are nullable

#### Organization Rules (1)
- `TestAlphabetize` - Ensures alphabetical ordering

#### Type Safety Rules (1)
- `TestListNonNullItems` - Ensures list items are non-null

#### Extensibility Rules (1)
- `TestEnumReservedValues` - Prevents reserved enum values

## Test Patterns

Each test follows a consistent pattern:

```go
func TestRuleName(t *testing.T) {
    rule := NewRuleName()

    t.Run("should flag violations", func(t *testing.T) {
        schema := `...schema with violations...`
        errors := runRule(t, rule, schema)
        if countRuleErrors(errors, "rule-name") == 0 {
            t.Error("Expected error for violation")
        }
    })

    t.Run("should pass valid schemas", func(t *testing.T) {
        schema := `...valid schema...`
        errors := runRule(t, rule, schema)
        if countRuleErrors(errors, "rule-name") > 0 {
            t.Error("Expected no errors for valid schema")
        }
    })
}
```

## Helper Functions

- `parseSchema(t, schemaStr)` - Parses GraphQL schema from string
- `runRule(t, rule, schemaStr)` - Runs a rule against a schema
- `containsError(errors, message)` - Checks if specific error exists
- `countRuleErrors(errors, ruleName)` - Counts errors for specific rule

## Test Schema

The `testdata/test_schema.graphql` file contains a comprehensive schema with intentional violations for manual testing:

```bash
go run . pkg/rules/testdata/test_schema.graphql
```

This will output all detected violations across all rules, demonstrating the linter's capabilities. 