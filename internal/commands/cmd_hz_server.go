package commands

import "github.com/spf13/cobra"

func NewHzServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Generate server code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	setupHzServerCommandFlags(cmd)
	return cmd
}

func setupHzServerCommandFlags(cmd *cobra.Command) {

}
