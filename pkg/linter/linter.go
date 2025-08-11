package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/anirudhraja/gqllinter/pkg/rules"
	"github.com/anirudhraja/gqllinter/pkg/types"
)

// Linter provides GraphQL schema linting functionality
type Linter struct {
	rules        []types.Rule
	enabledRules map[string]bool
}

// New creates a new linter instance with all built-in rules
func New() *Linter {
	return &Linter{
		rules: []types.Rule{
			// Yelp guidelines rules
			rules.NewTypesHaveDescriptions(),
			rules.NewFieldsHaveDescriptions(),
			rules.NewNoHashtagDescription(),
			rules.NewNamingConvention(),
			rules.NewNoFieldNamespacing(),
			rules.NewMinimalTopLevelQueries(),

			// Guild-inspired rules
			rules.NewNoUnusedFields(),
			rules.NewRequireDeprecationReason(),
			rules.NewNoScalarResultTypeOnMutation(),
			rules.NewAlphabetize(),
			rules.NewInputName(),

			// Additional comprehensive rules
			rules.NewNoUnusedTypes(),
			rules.NewCapitalizedDescriptions(),
			rules.NewEnumUnknownCase(),
			rules.NewNoQueryPrefixes(),
			rules.NewInputEnumSuffix(),
			rules.NewEnumDescriptions(),

			// Additional best practice rules
			rules.NewListNonNullItems(),
			rules.NewEnumReservedValues(),
			rules.NewMutationResponseNullable(),
			rules.NewOperationResponseName(),
			rules.NewFieldsNullableExceptId(),
			rules.NewRelayPageInfo(),
			rules.NewRelayEdgeTypes(),
			//rules.NewRelayConnectionTypes(), Fix the linting rules and then enable it later
		},
		enabledRules: make(map[string]bool),
	}
}

// LoadCustomRules loads custom rules from a directory containing Go plugins
func (l *Linter) LoadCustomRules(customRulesDir string) error {
	pluginFiles, err := filepath.Glob(filepath.Join(customRulesDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to find plugin files: %w", err)
	}

	for _, pluginFile := range pluginFiles {
		if err := l.loadPlugin(pluginFile); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", pluginFile, err)
		}
	}

	return nil
}

// loadPlugin loads a single Go plugin file
func (l *Linter) loadPlugin(pluginPath string) error {
	// Load the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for the NewRule function
	newRuleSymbol, err := p.Lookup("NewRule")
	if err != nil {
		return fmt.Errorf("plugin must export NewRule function: %w", err)
	}

	// Cast to expected function signature
	newRuleFunc, ok := newRuleSymbol.(func() types.Rule)
	if !ok {
		return fmt.Errorf("NewRule must be a function that returns types.Rule")
	}

	// Create the rule and add it to our rules list
	rule := newRuleFunc()
	l.rules = append(l.rules, rule)

	return nil
}

// SetRules enables only the specified rules
func (l *Linter) SetRules(ruleNames []string) {
	l.enabledRules = make(map[string]bool)
	for _, name := range ruleNames {
		l.enabledRules[name] = true
	}
}

// LintFile lints a single GraphQL schema file
func (l *Linter) LintFile(filename string) ([]types.LintError, error) {
	// Read and parse the schema
	schema, source, err := l.parseSchemaFile(filename)
	if err != nil {
		return nil, err
	}

	// Run all enabled rules
	var errors []types.LintError
	for _, rule := range l.rules {
		// Skip rule if specific rules are set and this rule is not enabled
		if len(l.enabledRules) > 0 && !l.enabledRules[rule.Name()] {
			continue
		}

		ruleErrors := rule.Check(schema, source)
		errors = append(errors, ruleErrors...)
	}

	return errors, nil
}

// parseSchemaFile reads and parses a GraphQL schema file
func (l *Linter) parseSchemaFile(filename string) (*ast.Schema, *ast.Source, error) {
	// Read file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Create source from file
	source := &ast.Source{
		Name:  filename,
		Input: string(content),
	}

	// Parse the schema
	schema, err := gqlparser.LoadSchema(source)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return schema, source, nil
}

// GetAvailableRules returns all available rule names
func (l *Linter) GetAvailableRules() []string {
	var ruleNames []string
	for _, rule := range l.rules {
		ruleNames = append(ruleNames, rule.Name())
	}
	return ruleNames
}
