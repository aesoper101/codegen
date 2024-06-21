package commands

import (
	"github.com/aesoper101/codegen/internal/generator/kitex"
	"github.com/spf13/cobra"
)

func NewKxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kx",
		Short:   "a wrapper for kitex command",
		PreRunE: envCheckPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			configFiles, err := cmd.Flags().GetStringSlice("config")
			if err != nil {
				return err
			}
			gen, err := kitex.NewGenerator(cmd.Context(), configFiles)
			if err != nil {
				return err
			}

			return gen.Generate()
		},
	}

	setupKitexCommandFlags(cmd)

	return cmd
}

func setupKitexCommandFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP("config", "c", []string{"kitex.yaml"}, "config file path")
}
