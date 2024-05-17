package commands

import "github.com/spf13/cobra"

func NewHzClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Generate client code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	setupHzClientCommandFlags(cmd)

	return cmd
}

func setupHzClientCommandFlags(cmd *cobra.Command) {

}
