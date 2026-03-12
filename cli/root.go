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
			result, err := ops.GenerateChangelog(ops.Input{
				BaseBranch:    args[0],
				FeatureBranch: args[1],
			})
			if err != nil {
				fmt.Printf(" %s %s\n", fail, err)
				return err
			}

			if !result.HasChanges {
				fmt.Printf(" %s No API changes between %s and %s\n", ok, args[0], args[1])
				return nil
			}

			if outputFile != "" {
				if prepend {
					existing, _ := os.ReadFile(outputFile)
					content := result.Markdown + "\n" + string(existing)
					if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
						return fmt.Errorf("writing %s: %w", outputFile, err)
					}
				} else {
					if err := os.WriteFile(outputFile, []byte(result.Markdown), 0644); err != nil {
						return fmt.Errorf("writing %s: %w", outputFile, err)
					}
				}
				fmt.Printf(" %s Changelog written to %s\n", ok, dim.Render(outputFile))
			} else {
				fmt.Print(result.Markdown)
			}

			return nil
		},
	}

	root.Flags().StringVarP(&outputFile, "output", "o", "", "Write changelog to file")
	root.Flags().BoolVar(&prepend, "prepend", false, "Prepend to existing file (requires --output)")

	return root
}
