package assembler

import (
	"testing"

	"github.com/leon332157/risc-y-8/cmd/r8/internal/assembler/grammar"
)

func TestOneOperand(t *testing.T) {
	var test = []struct {
		instr    grammar.Instruction
		expected uint32
	}{
		{
			instr: grammar.Instruction{
				Mnemonic: "nop",
				Operands: []grammar.Operand{},
			},
			expected: 0b00000000000000000001110000000101,
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "hlt",
				Operands: []grammar.Operand{},
			},
			expected: 0b11111111111111111110000000001101,
		},
		{instr: grammar.Instruction{
			Mnemonic: "ret",
			Operands: []grammar.Operand{},
		},
			expected: 0b000000000000001110000111111101,
		},
	}
	for num, testCase := range test {
		inst, err := parseOneInst(&testCase.instr)
		if err != nil {
			t.Fatalf("Failed to parse instruction %v: %v", num, err)
		}
		if inst.Encode() != testCase.expected {
			t.Errorf("Num: %v Expected %x, got %x", num, testCase.expected, inst.Encode())
		}
	}
}
