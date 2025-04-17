package r8

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "r8 [command]",
		Short: "risch-y-8",
		Long:  "risc-y-8 is a cpu architecture with toolchain"}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
