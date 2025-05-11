package r8

import (
	"fmt"
	"os"

	"encoding/binary"
	"github.com/leon332157/risc-y-8/cmd/r8/assembler"
	"github.com/leon332157/risc-y-8/cmd/r8/assembler/grammar"
	"github.com/spf13/cobra"
)

var (
	assembleCmd = &cobra.Command{
		Use:     "assemble <flags> [input file]",
		Aliases: []string{"as", "as"},
		Short:   "Assemble RISC-Y-8 assembly code",
		Long:    "Assemble RISC-Y-8 assembly code into machine code",
		RunE:    runAssemble,
		Args:    cobra.ExactArgs(1),
		Example: "r8 assemble -o a.out -f bin input.asm",
	}
)

func runAssemble(cmd *cobra.Command, args []string) error {
	infile := args[0]
	outfile := cmd.Flag("output").Value.String()
	//format := cmd.Flag("format").Value.String()
	f, err := os.Open(infile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer f.Close()
	if infile == "-" {
		return fmt.Errorf("Stdin is not supported yet")
	}

	prog, err := grammar.Parser.Parse(infile, f)
	if err != nil {
		return fmt.Errorf("parse file %v %+v",err,prog);
	}
	res,err := assembler.ParseLines(prog.Lines);
	if err != nil {
		return fmt.Errorf("parse lines: %v %+v", err, res)
	}
	encoded := assembler.EncInstructions(res)
	if outfile != "" {
		of, err := os.Create(outfile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer of.Close()
		binary.Write(of, binary.LittleEndian, encoded)

	} else {
		binary.Write(os.Stdout, binary.LittleEndian, encoded)
	}
	return nil
}

func init() {
	assembleCmd.Flags().StringP("output", "o", "", "Output machine code file")
	assembleCmd.Flags().StringP("format", "f", "bin", "Output format (bin, hex)")
	assembleCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	assembleCmd.Flags().MarkHidden("verbose") // Hide the verbose flag for now
	assembleCmd.Flags().MarkHidden("format")  // Hide the format flag for now

	rootCmd.AddCommand(assembleCmd)
	// Add flags and configuration settings here if needed
	// For example: assembleCmd.Flags().StringP("output", "o", "", "Output file")
}
