package cli

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/CheeziCrew/gruyere/internal/ops"
)

var (
	ok   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✔")
	fail = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗")
	dim  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// generateChangelogFn is the function used to generate changelogs. Replaceable in tests.
var generateChangelogFn = func(input ops.Input) (ops.Output, error) {
	return ops.GenerateChangelog(input)
}

// BuildCLI creates the Cobra command for gruyere.
func BuildCLI() *cobra.Command {
	var outputFile string
	var prepend bool

	root := &cobra.Command{
		Use:           "gruyere <base-branch> <feature-branch>",
		Short:         "Gruyere — diff OpenAPI specs between branches",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChangelog(args[0], args[1], outputFile, prepend)
		},
	}

	root.Flags().StringVarP(&outputFile, "output", "o", "", "Write changelog to file")
	root.Flags().BoolVar(&prepend, "prepend", false, "Prepend to existing file (requires --output)")

	return root
}

func runChangelog(baseBranch, featureBranch, outputFile string, prepend bool) error {
	result, err := generateChangelogFn(ops.Input{
		BaseBranch:    baseBranch,
		FeatureBranch: featureBranch,
	})
	if err != nil {
		fmt.Printf(" %s %s\n", fail, err)
		return err
	}

	if !result.HasChanges {
		fmt.Printf(" %s No API changes between %s and %s\n", ok, baseBranch, featureBranch)
		return nil
	}

	if outputFile == "" {
		fmt.Print(result.Markdown)
		return nil
	}

	return writeOutput(result.Markdown, outputFile, prepend)
}

func writeOutput(markdown, outputFile string, prepend bool) error {
	if prepend {
		existing, _ := os.ReadFile(outputFile)
		markdown = markdown + "\n" + string(existing)
	}
	if err := os.WriteFile(outputFile, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outputFile, err)
	}
	fmt.Printf(" %s Changelog written to %s\n", ok, dim.Render(outputFile))
	return nil
}
