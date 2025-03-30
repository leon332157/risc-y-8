package assembler

import (
	"github.com/leon332157/risc-y-8/cmd/r8/internal/assembler/grammar"
	. "github.com/leon332157/risc-y-8/pkg/types"
	"testing"
)

func TestEncodeBaseInstructions(t *testing.T) {

	// [0010 0011 0100 0101] 000 1100 01011 00 01
	regimm1 := BaseInstruction{			// LDI r11, 0x2345
		OpType: RegImm,					// 00
		Rd:     11,						// 01011
		ALU:    ImmALU["ldi"],			// 1100
		Imm:    0x2345,					// 0010 0011 0100 0101
	}

	// 0000000000010000 000 0001 00100 00 01
	regimm2 := BaseInstruction{			// SUB r4, 16
		OpType: RegImm,					// 00
		Rd:     4,						// 00100
		ALU:    ImmALU["sub"],			// 0001
		Imm:    16,						// 0000 0000 0001 0000
	}

	// 0000000000000001 000 0010 01100 00 01
	regimm3 := BaseInstruction{			// MUL r12, 0x1
		OpType: RegImm,					// 00
		Rd:     12,						// 01100
		ALU:    ImmALU["mul"],			// 0010
		Imm:    1,						// 0000 0000 0000 0001
	}

	// 00000000000000 00010 0000 00001 01 01
	regreg1 := BaseInstruction{			// ADD r1, r2
		OpType: RegReg,					// 01
		Rd:     1,						// 00001
		ALU:    0000,					// 0000
		Rs:     2,						// 00010
	}

	// 00000000000000 00101 0100 00100 01 01
	regreg2 := BaseInstruction{			// REM r4, r5
		OpType: RegReg,					// 01
		Rd:     4,						// 00100
		ALU:    RegALU["rem"],			// 0100
		Rs:     5,						// 00101
	}

	// 00000000000000 01010 1110 01011 01 01
	regreg3 := BaseInstruction{			// CPY r11, r10
		OpType: RegReg,					// 01
		Rd:     11,						// 01011
		ALU:	RegALU["cpy"],			// 1110
		Rs:     10,						// 01010
	}

	// 0000000100000000 00100 00 00011 10 01
	ldstr1 := BaseInstruction{			// LDW r3, [r4 + 0x100]
		OpType: LoadStore,				// 10
		Rd:     3,						// 00011
		Mode:   LDW,					// 00
		RMem:   4,						// 00100
		Imm:    0x100,					// 0000 0001 0000 0000
	}

	// 0000000011111111 00011 11 00010 10 01
	ldstr2 := BaseInstruction{			// STW r2, [r3 + 0xff]
		OpType: LoadStore,				// 10
		Rd:     2,						// 00010
		Mode:   STW,					// 11
		RMem:   3,						// 00011
		Imm:    0xff,					// 0000 0000 1111 1111
	}

	// 000000000000000000000 10 00100 10 01
	ldstr3 := BaseInstruction{			// PUSH r4
		OpType: LoadStore,				// 10
		Rd:     4,						// 00100
		Mode:   PUSH,					// 10
	}

	//  0000001000000000 000 0001 00101 11 01
	ctrl1 := BaseInstruction{			// BEQ [r5 + 0x200]
		OpType: Control,				// 11
		RMem:   5,						// 00101
		Flag:   Conditions["eq"].Flag,	// BEQ (Flag 0001)
		Mode:   Conditions["eq"].Mode,	// 000
		Imm:    0x200,					// 0000 0010 0000 0000
	}

	// 0000010000000000 100 0100 00110 11 01
	ctrl2 := BaseInstruction{			// BOF [r6 + 0x400]
		OpType: Control,				// 11
		RMem:   6,						// 00110
		Flag:   Conditions["of"].Flag,	// BOF (Flag 0100)
		Mode:   Conditions["of"].Mode,	// 100
		Imm:    0x400,					// 0000 0100 0000 0000
	}

	// 0000000000000000 000 0100 01010 11 01
	ctrl3 := BaseInstruction{			// BNF r10
		OpType: Control,				// 11
		RMem:   10,						// 01010 (r10)
		Flag:   Conditions["nf"].Flag,	// BNF (Flag 0100)
		Mode:   Conditions["nf"].Mode,	// 000
	}

	tests := []struct {
		name     string
		inst     BaseInstruction
		expected uint32
	}{
		{"regimm1", regimm1, 0b00100011010001010001100010110001},
		{"regimm2", regimm2, 0b00000000000100000000001001000001},
		{"regimm3", regimm3, 0b00000000000000010000010011000001},
		
		{"regreg1", regreg1, 0b00000000000000000100000000010101},
		{"regreg2", regreg2, 0b00000000000000001010100001000101},
		{"regreg3", regreg3, 0b00000000000000010101110010110101},

		{"ldstr1", ldstr1, 0b00000001000000000010000000111001},
		{"ldstr2", ldstr2, 0b00000000111111110001111000101001},
		{"ldstr3", ldstr3, 0b00000000000000000000010001001001},

		{"ctrl1", ctrl1, 0b00000010000000000000001001011101},
		{"ctrl2", ctrl2, 0b00000100000000001000100001101101},
		{"ctrl3", ctrl3, 0b00000000000000000000100010101101},
	}

	for _, tt := range tests {

		if encoded := tt.inst.Encode(); encoded != tt.expected {
			t.Errorf("%s: expected 0x%08X, got 0x%08X", tt.name, tt.expected, encoded)
		}
	}
}

func TestEncodeBaseNoOperand(t *testing.T) {
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

func TestDecodeBaseInstruction(t *testing.T) {

	var test = []struct {
		encoded  uint32
		expected BaseInstruction
	}{
		{
			encoded: 0b00100011010001010001100010110001, // LDI r11, 0x2345
			expected: BaseInstruction{
				OpType: RegImm,							// 00
				Rd:     11,								// 01011
				ALU:    ImmALU["ldi"],					// 1100
				Imm:    0x2345,							// 0010 0011 0100 0101
			},
		},
		{
			encoded: 0b00000000000100000000001001000001, // SUB r4, 16
			expected: BaseInstruction{
				OpType: RegImm,							// 00
				Rd:     4,								// 00100
				ALU:    ImmALU["sub"],					// 0001
				Imm:    16,								// 0000 0000 0001 0000
			},
		},
		{
			encoded: 0b00000000000000010000010011000001, // MUL r12, 0x1
			expected: BaseInstruction{
				OpType: RegImm,							// 00
				Rd:     12,								// 01100
				ALU:    ImmALU["mul"],					// 0010
				Imm:    1,								// 0000 0000 0000 0001
			},
		},
		{
			encoded: 0b00000000000000000100000000010101,// ADD r1, r2
			expected: BaseInstruction{
				OpType: RegReg, 						// 01
				Rd:     1,      						// 00001
				ALU:    0,      						// 0000 (ADD operation)
				Rs:     2,      						// 00010
			},
		},
		{
			encoded: 0b00000000000000001010100001000101,// REM r4, r5
			expected: BaseInstruction{
				OpType: RegReg,							// 01
				Rd:     4,								// 00100
				ALU:    RegALU["rem"],					// 0100
				Rs:     5,								// 00101
			},
		},
		{
			encoded: 0b00000000000000010101110010110101,// CPY r11, r10
			expected: BaseInstruction{
				OpType: RegReg,							// 01
				Rd:     11,								// 01011
				ALU:    RegALU["cpy"],					// 1110
				Rs:     10,								// 01010
			},
		},
		{
			encoded: 0b00000001000000000010000000111001,// LDW r3, [r4 + 0x100]
			expected: BaseInstruction{
				OpType: LoadStore,						// 10
				Rd:     3,								// 00011
				Mode:   LDW,							// 00 (Load Word)
				RMem:   4,								// 00100 (r4)
				Imm:    0x100,							// 0000 0001 0000 0000
			},
		},
		{
			encoded: 0b00000000111111110001111000101001, // STW r2, [r3 + 0xff]
			expected: BaseInstruction{
				OpType: LoadStore,						// 10
				Rd:     2,								// 00010
				Mode:   STW,							// 11 (Store Word)
				RMem:   3,								// 00011 (r3)
				Imm:    0xff,							// 0000 0000 1111 1111
			},
		},
		{
			encoded: 0b00000000000000000000010001001001,// PUSH r4
			expected: BaseInstruction{
				OpType: LoadStore, 						// 10
				Rd:     4,								// 00100
				Mode:   PUSH,							// 10 (Push)
			},
		},
		{
			encoded: 0b00000010000000000000001001011101,// BEQ [r5 + 0x200]
			expected: BaseInstruction{
				OpType: Control,						// 11
				RMem:   5,								// 00101 (r5)
				Flag:   Conditions["eq"].Flag,			// BEQ (Flag 0001)
				Mode:   Conditions["eq"].Mode,			// 000 (Normal mode)
				Imm:    0x200,							// 0000 0010 0000 0000
			},
		},
		{
			encoded: 0b00000100000000001000100001101101,// BOF [r6 + 0x400]
			expected: BaseInstruction{
				OpType: Control,						// 11
				RMem:   6,								// 00110 (r6)
				Flag:   Conditions["of"].Flag,			// BOF (Flag 0100)
				Mode:   Conditions["of"].Mode,			// 100 (Normal mode)
				Imm:    0x400,							// 0000 0100 0000 0000
			},
		},
		{
			encoded: 0b00000000000000000000100010101101,// BNF r10
			expected: BaseInstruction{
				OpType: Control,						// 11
				RMem:   10,								// 01010 (r10)
				Flag:   Conditions["nf"].Flag,			// BNF (Flag 0100)
				Mode:   Conditions["nf"].Mode,			// 000 (Normal mode)
				Imm:    0,								// 0000 0000 0000 0000
			},
		},

	}

	for num, testCase := range test {
		var inst BaseInstruction
		inst.Decode(testCase.encoded)

		// Compare the decoded instruction with the expected instruction
		if inst.OpType != testCase.expected.OpType ||
			inst.Rd != testCase.expected.Rd ||
			inst.ALU != testCase.expected.ALU ||
			inst.Rs != testCase.expected.Rs ||
			inst.RMem != testCase.expected.RMem ||
			inst.Flag != testCase.expected.Flag ||
			inst.Mode != testCase.expected.Mode ||
			inst.Imm != testCase.expected.Imm {
			t.Errorf("Num: %v, expected: %+v, got: %+v", num, testCase.expected, inst)
		}
	}
}