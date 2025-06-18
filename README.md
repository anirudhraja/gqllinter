# GraphQL Schema Linter

A comprehensive GraphQL schema linter that helps enforce schema design best practices, based on [Yelp's GraphQL guidelines](https://yelp.github.io/graphql-guidelines/schema-design.html) and other industry standards.

## Features

- **Extensible Rule System**: Implement custom rules by satisfying the `Rule` interface
- **Multiple Output Formats**: Support for both text and JSON output formats
- **Glob Pattern Support**: Lint multiple files using glob patterns
- **Built-in Rules**: Comprehensive set of rules following industry best practices
- **Command Line Interface**: Easy-to-use CLI similar to existing GraphQL linters

## Installation

```bash
go install github.com/gqllinter@latest
```

Or build from source:

```bash
git clone https://github.com/gqllinter/gqllinter.git
cd gqllinter
go build -o gqllinter
```

## Usage

### Basic Usage

```bash
# Lint a single file
gqllinter schema.graphql

# Lint multiple files with glob patterns
gqllinter schema/*.graphql

# Output as JSON
gqllinter --format json schema.graphql

# Save output to file
gqllinter --format json --output results.json schema.graphql

# Run only specific rules
gqllinter --rules types-have-descriptions,fields-have-descriptions schema.graphql
```

### Command Line Options

```
Usage:
  gqllinter [flags] <schema-files>

Flags:
      --config string              path to configuration file
      --custom-rule-paths string   path to custom rules directory
      --format string              output format (text, json) (default "text")
      --ignore string              comment to ignore linting errors (default "# gqllinter-ignore")
      --output string              output file (default: stdout)
      --rules strings              comma-separated list of rules to run
```

## Built-in Rules

### types-have-descriptions
Ensures all types have descriptions to explain their purpose.

**Bad:**
```graphql
type User {
  id: ID!
  name: String!
}
```

**Good:**
```graphql
"""
Represents a user in the system
"""
type User {
  id: ID!
  name: String!
}
```

### fields-have-descriptions
Ensures all fields have descriptions to explain their purpose.

**Bad:**
```graphql
type User {
  id: ID!
  name: String!
}
```

**Good:**
```graphql
type User {
  """Unique identifier for the user"""
  id: ID!
  
  """Full name of the user"""
  name: String!
}
```

### no-hashtag-description
Enforces the use of triple quotes for descriptions instead of hashtag comments, following Yelp guidelines.

**Bad:**
```graphql
# Represents a user in the system
type User {
  id: ID!
}
```

**Good:**
```graphql
"""
Represents a user in the system
"""
type User {
  id: ID!
}
```

### naming-convention
Enforces specific naming conventions and discourages generic type names.

**Bad:**
```graphql
type Data {
  info: String
}
```

**Good:**
```graphql
type UserData {
  userInfo: String
}
```

## Custom Rules

You can create custom rules by implementing the `Rule` interface:

```go
package main

import (
    "github.com/vektah/gqlparser/v2/ast"
    "github.com/gqllinter/pkg/linter"
)

type MyCustomRule struct{}

func (r *MyCustomRule) Name() string {
    return "my-custom-rule"
}

func (r *MyCustomRule) Description() string {
    return "Description of what this rule checks"
}

func (r *MyCustomRule) Check(schema *ast.Schema, source *ast.Source) []linter.LintError {
    var errors []linter.LintError
    
    // Your custom validation logic here
    
    return errors
}

// For plugins, export this function
func NewRule() linter.Rule {
    return &MyCustomRule{}
}
```

Compile your custom rule as a plugin:

```bash
go build -buildmode=plugin -o my-rule.so my-rule.go
```

Then use it with the linter:

```bash
gqllinter --custom-rule-paths ./rules/ schema.graphql
```

## Output Format

### Text Format (Default)

```
schema.graphql:5:1: The object type `QueryRoot` is missing a description. (types-have-descriptions)
schema.graphql:6:3: The field `QueryRoot.a` is missing a description. (fields-have-descriptions)
```

### JSON Format

```json
{
  "errors": [
    {
      "message": "The object type `QueryRoot` is missing a description.",
      "location": {
        "line": 5,
        "column": 1,
        "file": "schema.graphql"
      },
      "rule": "types-have-descriptions"
    },
    {
      "message": "The field `QueryRoot.a` is missing a description.",
      "location": {
        "line": 6,
        "column": 3,
        "file": "schema.graphql"
      },
      "rule": "fields-have-descriptions"
    }
  ]
}
```

## Configuration

Create a configuration file to customize the linter behavior:

```yaml
# .gqllinter.yml
rules:
  - types-have-descriptions
  - fields-have-descriptions
  - naming-convention

ignore-patterns:
  - "# gqllinter-ignore"

custom-rules-dir: "./custom-rules"
```

## Integration

### GitHub Actions

```yaml
name: GraphQL Schema Lint
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Install gqllinter
        run: go install github.com/gqllinter@latest
      - name: Lint GraphQL Schema
        run: gqllinter --format json schema/*.graphql
```

### Pre-commit Hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: gqllinter
        name: GraphQL Schema Lint
        entry: gqllinter
        language: system
        files: \.graphql$
```

## References

This linter is inspired by and follows guidelines from:

- [Yelp GraphQL Guidelines](https://yelp.github.io/graphql-guidelines/schema-design.html)
- [The Guild GraphQL ESLint Rules](https://the-guild.dev/graphql/eslint/rules)
- [graphql-schema-linter](https://github.com/cjoudrey/graphql-schema-linter)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 
