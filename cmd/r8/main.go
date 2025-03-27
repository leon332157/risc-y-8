package r8

import (
	"fmt"
	"github.com/leon332157/risc-y-8/cmd/r8/internal/assembler"
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

func main1() {
	fmt.Println("hello from main")
	//assembler.ParseString("add r1, r2\nadd r3, -1")
	fmt.Printf("%x\n", assembler.EncInstructions())
}
