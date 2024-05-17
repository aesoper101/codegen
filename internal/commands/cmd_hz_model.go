package commands

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"os"
)

func NewHzModelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "model",
		Short:   "Generate model code",
		PreRunE: envCheckPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			skip, _ := cmd.Flags().GetBool("yes")
			input, _ := cmd.Flags().GetString("input")
			if input == "" && !skip {
				if err := huh.NewInput().Title("Please provide an input directory").Validate(func(s string) error {
					stat, err := os.Stat(s)
					if err != nil {
						return err
					}
					if !stat.IsDir() {
						return fmt.Errorf("%s is not a directory", s)
					}
					return nil
				}).Value(&input).Run(); err != nil {
					return err
				}
			} else if input == "" {
				input = "idl"
			}

			outDir, _ := cmd.Flags().GetString("out")
			if outDir == "" && !skip {
				if err := huh.NewInput().Title("Please provide an output directory").Value(&outDir).Run(); err != nil {
					return err
				}
			} else if outDir == "" {
				outDir = "model"
			}
			return nil
		},
	}

	setupHzModelCommandFlags(cmd)

	return cmd
}

func setupHzModelCommandFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("input", "i", "idl", "the directory containing the spec file")
	cmd.PersistentFlags().StringP("out", "o", "model", "output directory for generated code")
	cmd.PersistentFlags().BoolP("yes", "y", false, "skip interactive prompts")
}
