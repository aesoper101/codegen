package commands

import (
	"github.com/aesoper101/codegen/internal/generator/hertz"
	"github.com/spf13/cobra"
)

func NewHzCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "hz",
		Short:   "a wrapper for hertz command",
		PreRunE: envCheckPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFiles, err := cmd.Flags().GetStringSlice("config")
			if err != nil {
				return err
			}
			gen, err := hertz.NewGenerator(cmd.Context(), configFiles)
			if err != nil {
				return err
			}

			return gen.Generate()
		},
	}

	setupHzCommandFlags(cmd)

	return cmd
}

func setupHzCommandFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP("config", "c", []string{"hertz.yaml"}, "config file path")
}
