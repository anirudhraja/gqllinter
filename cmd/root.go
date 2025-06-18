package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/linter"
	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/spf13/cobra"
)

var (
	configFile     string
	format         string
	outputFile     string
	rules          []string
	ignorePragma   string
	customRulesDir string
)

var rootCmd = &cobra.Command{
	Use:   "gqllinter [flags] <schema-files>",
	Short: "GraphQL schema linter following Yelp guidelines",
	Long: `A GraphQL schema linter that helps enforce schema design best practices
based on Yelp's GraphQL guidelines and other industry standards.

Examples:
  gqllinter schema.graphql
  gqllinter --format json --output results.json schema/*.graphql
  gqllinter --rules types-have-descriptions,fields-have-descriptions schema.graphql`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLint,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "path to configuration file")
	rootCmd.PersistentFlags().StringVar(&format, "format", "text", "output format (text, json)")
	rootCmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file (default: stdout)")
	rootCmd.PersistentFlags().StringSliceVar(&rules, "rules", []string{}, "comma-separated list of rules to run")
	rootCmd.PersistentFlags().StringVar(&ignorePragma, "ignore", "# gqllinter-ignore", "comment to ignore linting errors")
	rootCmd.PersistentFlags().StringVar(&customRulesDir, "custom-rule-paths", "", "path to custom rules directory")
}

func runLint(cmd *cobra.Command, args []string) error {
	// Expand glob patterns in arguments
	var schemaFiles []string
	for _, pattern := range args {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
		}
		schemaFiles = append(schemaFiles, matches...)
	}

	if len(schemaFiles) == 0 {
		return fmt.Errorf("no schema files found")
	}

	// Create linter instance
	l := linter.New()

	// Load custom rules if specified
	if customRulesDir != "" {
		if err := l.LoadCustomRules(customRulesDir); err != nil {
			return fmt.Errorf("failed to load custom rules: %w", err)
		}
	}

	// Set specific rules if provided
	if len(rules) > 0 {
		l.SetRules(rules)
	}

	// Lint all schema files
	var allErrors []types.LintError
	for _, file := range schemaFiles {
		errors, err := l.LintFile(file)
		if err != nil {
			return fmt.Errorf("failed to lint %s: %w", file, err)
		}
		allErrors = append(allErrors, errors...)
	}

	// Output results
	return outputResults(allErrors)
}

func outputResults(errors []types.LintError) error {
	var output string
	var err error

	switch format {
	case "json":
		output, err = formatJSON(errors)
	case "text":
		output = formatText(errors)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return err
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(output), 0644)
	}

	fmt.Print(output)
	return nil
}

func formatJSON(errors []types.LintError) (string, error) {
	result := struct {
		Errors []types.LintError `json:"errors"`
	}{
		Errors: errors,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func formatText(errors []types.LintError) string {
	if len(errors) == 0 {
		return "No linting errors found.\n"
	}

	var lines []string
	for _, err := range errors {
		line := fmt.Sprintf("%s:%d:%d: %s (%s)",
			err.Location.File,
			err.Location.Line,
			err.Location.Column,
			err.Message,
			err.Rule,
		)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n") + "\n"
}
