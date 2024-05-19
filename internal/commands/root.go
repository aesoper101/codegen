package commands

import (
	"github.com/spf13/cobra"
)

func NewCodegenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codegen",
		Short: "Generate code from a spec file",
	}

	setupCodegenCommandFlags(cmd)

	kxCmd := NewKxCommand()

	hzCmd := NewHzCommand()
	hzCmd.AddCommand(NewHzModelCommand())
	hzCmd.AddCommand(NewHzClientCommand())
	hzCmd.AddCommand(NewHzServerCommand())

	cmd.AddCommand(kxCmd, hzCmd)
	return cmd
}

func setupCodegenCommandFlags(cmd *cobra.Command) {
}
