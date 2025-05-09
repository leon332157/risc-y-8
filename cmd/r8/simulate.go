package r8

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/leon332157/risc-y-8/cmd/r8/simulator"
	"github.com/rs/zerolog"
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
	simulateCmd.Flags().BoolVar(&disableCache, "disable-cache", false, "Disable cache")
	simulateCmd.Flags().BoolVar(&disablePipeline, "disable-pipeline", false, "Disable pipeline")
	rootCmd.AddCommand(simulateCmd)
}

func runSimulate(cmd *cobra.Command, args []string) error {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	infile := args[0]
	f, err := os.ReadFile(infile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	program := make([]uint32, len(f)/4)
	bytesReader := bytes.NewReader(f)
	err = binary.Read(bytesReader, binary.LittleEndian, &program)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}
	sys := simulator.NewSystem(program, disableCache, disablePipeline)
	sys.RunToEnd(nil)
	return nil
}
