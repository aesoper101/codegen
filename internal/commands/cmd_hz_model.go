package commands

import (
	hzmodel "github.com/aesoper101/codegen/internal/generator/hertz/model"
	"github.com/spf13/cobra"
)

func NewHzModelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Generate model code",
		//PreRunE: envCheckPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := cmd.Flags().GetString("input")
			outDir, _ := cmd.Flags().GetString("out")

			return hzmodel.New(
				hzmodel.WithIDLPath(input),
				hzmodel.WithOutPath(outDir),
			).Generate()
		},
	}

	setupHzModelCommandFlags(cmd)

	return cmd
}

func setupHzModelCommandFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("input", "i", "idl", "the directory containing the spec file")
	cmd.PersistentFlags().StringP("out", "o", "model", "output directory for generated code")
}
