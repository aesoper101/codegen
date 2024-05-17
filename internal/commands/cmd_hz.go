package commands

import "github.com/spf13/cobra"

func NewHzCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hz",
		Short: "a wrapper for hertz command",
	}

	setupHzCommandFlags(cmd)

	return cmd
}

func setupHzCommandFlags(cmd *cobra.Command) {}
