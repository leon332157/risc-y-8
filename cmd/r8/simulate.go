package r8

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/leon332157/risc-y-8/cmd/r8/simulator"
	"github.com/spf13/cobra"
)

var (
	simulateCmd = &cobra.Command{
		Use:     "simulate <flags> [binary file]",
		Aliases: []string{"sim"},
		Short:   "Simulate RISC-Y-8 binary",
		//Long:    "Assemble RISC-Y-8 assembly code into machine code",
		RunE:    runSimulate,
		Args:    cobra.ExactArgs(1),
		Example: "r8 simulate input.bin",
	}
)

func init() {
	rootCmd.AddCommand(simulateCmd)
}

func runSimulate(cmd *cobra.Command, args []string) error {
	infile := args[0]
	f, err := os.Open(infile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer f.Close()
	program := make([]uint32, 0)
	err = binary.Read(f, binary.LittleEndian, &program)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}
	sys := simulator.NewSystem(program, false, false)
	sys.RunForever(nil)
	return nil
}
