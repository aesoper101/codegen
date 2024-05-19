package commands

import (
	hzclient "github.com/aesoper101/codegen/internal/generator/hertz/client"
	"github.com/spf13/cobra"
)

func NewHzClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "client",
		Short:   "Generate client code",
		PreRunE: envCheckPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := cmd.Flags().GetString("input")
			outDir, _ := cmd.Flags().GetString("out")
			model, _ := cmd.Flags().GetString("model")
			baseDomain, _ := cmd.Flags().GetString("base_domain")
			opts := []hzclient.CliGeneratorOption{
				hzclient.WithOutPath(outDir),
				hzclient.WithIdlPath(input),
			}
			if model != "" {
				opts = append(opts, hzclient.WithModelMod(model))
			}
			if baseDomain != "" {
				opts = append(opts, hzclient.WithBaseDomain(baseDomain))
			}
			return hzclient.New(opts...).Generate()
		},
	}

	setupHzClientCommandFlags(cmd)

	return cmd
}

func setupHzClientCommandFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("input", "i", "idl", "the directory containing the spec file")
	cmd.PersistentFlags().StringP("model", "m", "", "the directory containing the spec file")
	cmd.PersistentFlags().StringP("out", "o", "client", "output directory for generated code")
	cmd.PersistentFlags().String("base_domain ", "", "Specify the request domain")
}
