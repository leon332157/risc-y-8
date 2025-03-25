package assembler

import (
	"github.com/leon332157/risc-y-8/cmd/r8/internal/assembler/grammar"
	"testing"
)

func TestEncodeInstructions(t *testing.T) {
	inst1 := BaseInstruction{
		// ADD r1, r2
		OpType: RegReg, // 01
		Rd:     1,      // 00001
		ALU:    0000,   // 0000
		Rs:     2,      // 00010
	}

	inst2 := BaseInstruction{
		// LDW r3, [r4 + 0x100]
		OpType: LoadStore, // 10
		Rd:     3,         // 00011
		Mode:   LDW,       // 00
		RMem:   4,         // 00100
		Imm:    0x100,     // 0000 0001 0000 0000
	}

	inst3 := BaseInstruction{
		// BEQ [r5 + 0x200]
		OpType: Control,               // 11
		RMem:   5,                     // 00101
		Flag:   Conditions["eq"].Flag, // BEQ (Flag 0000, Mode 000)
		Mode:   Conditions["eq"].Mode, // 000
		Imm:    0x200,                 // 0000 0000 0010 0000
	}

	inst4 := BaseInstruction{
		// BOF [r6 + 0x400]
		OpType: Control,
		Flag:   Conditions["of"].Flag, // BOF (Flag 0100, Mode 001)
		Mode:   Conditions["of"].Mode, // 001
		RMem:   6,                     // 00110
		Imm:    0x400,                 // 0000 0100 0000 0000
	}

	inst5 := BaseInstruction{
		// LDI r11, 0x2345
		OpType: RegImm,
		ALU:    ImmALU["ldi"], // LDI 1100
		Rd:     11,            // 01011
		Imm:    0x2345,        // 0010 0100 0110 1000 101
	}

	tests := []struct {
		name     string
		inst     BaseInstruction
		expected uint32
	}{
		{"inst1", inst1, 0b00000000000000000100000000010101},
		{"inst2", inst2, 0b00000001000000000010000000111001},
		{"inst3", inst3, 0b00000010000000000000001001011101},
		{"inst4", inst4, 0b00000100000000001000100001101101},
		{"inst5", inst5, 0b00100011010001010001100010110001},
	}

	for _, tt := range tests {

		if encoded := tt.inst.Encode(); encoded != tt.expected {
			t.Errorf("%s: expected 0x%08X, got 0x%08X", tt.name, tt.expected, encoded)
		}
	}
}

func TestEncodeNoOperand(t *testing.T) {
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
		inst, err := parseInst(&testCase.instr)
		if err != nil {
			t.Fatalf("Failed to parse instruction %v: %v", num, err)
		}
		if inst.Encode() != testCase.expected {
			t.Errorf("Num: %v Expected %x, got %x", num, testCase.expected, inst.Encode())
		}
	}
}
