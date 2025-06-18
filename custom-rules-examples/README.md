# Custom Rules Example

This directory demonstrates how to create and use custom rules with the GraphQL linter.

## How Custom Rules Work

Custom rules extend the linter's functionality by implementing the `types.Rule` interface. They are compiled as Go plugins (`.so` files) and loaded dynamically at runtime.

## Creating a Custom Rule

1. **Implement the Rule Interface**: Your rule must implement these methods:
   - `Name() string` - Returns the rule identifier
   - `Description() string` - Explains what the rule does  
   - `Check(schema *ast.Schema, source *ast.Source) []types.LintError` - Performs the actual linting

2. **Export a NewRule Function**: Your plugin must export a `NewRule()` function that returns your rule:
   ```go
   func NewRule() types.Rule {
       return &MyCustomRule{}
   }
   ```

3. **Build as Plugin**: Compile your rule as a Go plugin:
   ```bash
   go build -buildmode=plugin -o my-rule.so my-rule.go
   ```

## Example: Field ID Suffix Rule

The `field_id_suffix.go` file demonstrates a custom rule that ensures ID fields end with 'ID' not 'Id' for consistency.

### Building and Testing

```bash
# Build the plugin
go build -buildmode=plugin -o field_id_suffix.so field_id_suffix.go

# Test with the example schema
cd ../..
go run . --custom-rule-paths custom-rules-examples custom-rules-examples/test-schema.graphql

# Run only the custom rule
go run . --custom-rule-paths custom-rules-examples --rules field-id-suffix custom-rules-examples/test-schema.graphql

# Get JSON output
go run . --custom-rule-paths custom-rules-examples --rules field-id-suffix --format json custom-rules-examples/test-schema.graphql
```

### Expected Output

The custom rule should detect these issues in `test-schema.graphql`:
- `User.userId` should be `userID`
- `User.profileId` should be `profileID`  
- `CreateUserInput.profileId` should be `profileID`

## Rule Development Tips

1. **Skip Introspection Types**: Always skip GraphQL introspection types that start with `__`
2. **Handle Position Info**: Check if `field.Position` is not nil before accessing line/column
3. **Provide Helpful Messages**: Include specific suggestions for fixing the issue
4. **Test Thoroughly**: Test your rule with various schema patterns

## Available Rule Interface

```go
type Rule interface {
    Name() string
    Description() string
    Check(schema *ast.Schema, source *ast.Source) []types.LintError
}

type LintError struct {
    Message  string   `json:"message"`
    Location Location `json:"location"`
    Rule     string   `json:"rule"`
}

type Location struct {
    Line   int    `json:"line"`
    Column int    `json:"column"`
    File   string `json:"file"`
}
```

## Integration

Custom rules can be integrated into CI/CD pipelines, pre-commit hooks, and development workflows just like built-in rules. They support all the same output formats and command-line options. 