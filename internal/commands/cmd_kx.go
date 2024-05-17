package commands

import "github.com/spf13/cobra"

func NewKxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kx",
		Short: "a wrapper for kitex command",
	}

	setupKitexCommandFlags(cmd)

	return cmd
}

func setupKitexCommandFlags(cmd *cobra.Command) {

}
