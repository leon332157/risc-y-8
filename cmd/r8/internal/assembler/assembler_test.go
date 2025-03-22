package assembler

import (
	"testing"

	"github.com/leon332157/risc-y-8/cmd/r8/internal/assembler/grammar"
)

type testCase struct {
	instr         grammar.Instruction
	expected      BaseInstruction
	errorExpected bool
}

func runTests(t *testing.T, tests *[]testCase) {
	for num, testCase := range *tests {
		inst, err := parseInst(&testCase.instr)
		if err != nil {
			if testCase.errorExpected {
				t.Logf("Expected error for instruction %v: %+v %v", num, inst, err)
			} else {
				t.Fatalf("Failed test on instruction %v: %v, expected error", num, err)
			}
		}
		if inst != testCase.expected {
			t.Errorf("Num: %v Expected %+v, got %+v", num, testCase.expected, inst)
		}
	}
}

func makeImmediate(value string) grammar.OperandImmediate {
	return grammar.OperandImmediate{
		Value: value,
	}
}

func makeMemory(base, operation, displacement string) grammar.OperandMemory {
	return grammar.OperandMemory{
		Value: grammar.Memory{
			Base:      base,
			Operation: operation,
			Displacement: grammar.Displacement{
				Value: displacement,
			},
		},
	}
}

func makeRegister(value string) grammar.OperandRegister {
	return grammar.OperandRegister{
		Value: value,
	}
}

func makeOperands(operands ...grammar.Operand) []grammar.Operand {
	return operands
}

func TestPushPop(t *testing.T) {
	var tests = []testCase{{
		instr: grammar.Instruction{
			Mnemonic: "push",
			Operands: makeOperands(
				makeRegister("r1"))},

		expected: BaseInstruction{
			OpType: LoadStore,
			Rd:     1,
			Mode:   PUSH,
		},
	}, {
		instr: grammar.Instruction{
			Mnemonic: "pop",
			Operands: makeOperands(makeImmediate("0x100"))},
		expected:      BaseInstruction{},
		errorExpected: true,
	}, {
		instr: grammar.Instruction{
			Mnemonic: "pop",
			Operands: makeOperands(
				makeMemory("r1", "add", "0x100"))},

		expected:      BaseInstruction{},
		errorExpected: true,
	},
	}
	runTests(t, &tests)
}

func TestCallRet(t *testing.T) {
	var tests = []testCase{
		{
			instr: grammar.Instruction{
				Mnemonic: "call",
				Operands: makeOperands(
					makeMemory("", "", ""))},
			expected:      BaseInstruction{},
			errorExpected: true,
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "call",
				Operands: makeOperands(
					makeRegister("r1"))},
			expected:      BaseInstruction{},
			errorExpected: true,
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "call",
				Operands: makeOperands(makeMemory("r1", "+", "0xFFFFFF"))},
			expected:      BaseInstruction{},
			errorExpected: true,
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "call",
				Operands: makeOperands(makeMemory("r31", "+", "0x100"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   31,
				Mode:   0b111,
				Flag:   0b1111,
				Imm:    0x100,
			},
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "call",
				Operands: makeOperands(makeMemory("r31", "+", "-1"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   31,
				Mode:   0b111,
				Flag:   0b1111,
				Imm:    -1,
			},
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "call",
				Operands: makeOperands(makeMemory("r31", "-", "-1"))}, // can not have both operation and displacement be negative
			expected:      BaseInstruction{},
			errorExpected: true,
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "call",
				Operands: makeOperands(makeMemory("r10", "-", "0x100"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   10,
				Mode:   0b111,
				Flag:   0b1111,
				Imm:    -0x100,
			},
		},
	}
	runTests(t, &tests)

	tests = []testCase{
		{
			instr: grammar.Instruction{
				Mnemonic: "ret",
				Operands: makeOperands(makeRegister("r1"))},
			expected:      BaseInstruction{},
			errorExpected: true,
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "ret",
				Operands: makeOperands()},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   31,
				Mode:   0b111,
				Flag:   0b0000,
			},
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "ret",
				Operands: makeOperands(makeMemory("r1", "-", "0xFF"))},
			expected: BaseInstruction{}, errorExpected: true,
		},
	}
	runTests(t, &tests)
}

func TestBranch(t *testing.T) {
	var tests = []testCase{
		{
			instr: grammar.Instruction{
				Mnemonic: "bne",
				Operands: makeOperands(
					makeMemory("r1", "+", "0x111"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   1,
				Mode:   0b0,
				Flag:   0b0,
				Imm:    0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bz",
				Operands: makeOperands(
					makeMemory("r9", "+", "-10"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   9,
				Mode:   0b0,
				Flag:   0b1,
				Imm:    -10,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "blt",
				Operands: makeOperands(
					makeMemory("r29", "-", "0x111"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   29,
				Mode:   0b001,
				Flag:   0b0110,
				Imm:    -0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bge",
				Operands: makeOperands(
					makeMemory("r31", "-", "0x111"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   31,
				Mode:   0b011,
				Flag:   0b0110,
				Imm:    -0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "blu",
				Operands: makeOperands(
					makeMemory("r31", "-", "0x111"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   31,
				Mode:   0b100,
				Flag:   0b1000,
				Imm:    -0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bae",
				Operands: makeOperands(
					makeMemory("r31", "-", "0x111"))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   31,
				Mode:   0b000,
				Flag:   0b1000,
				Imm:    -0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "ba",
				Operands: makeOperands(
					makeMemory("r31", "", ""))},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   31,
				Mode:   0b010,
				Flag:   0b1000,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bof",
				Operands: makeOperands(
					makeMemory("r0", "", "")),
			},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   0,
				Mode:   0b100,
				Flag:   0b0100,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bnf",
				Operands: makeOperands(
					makeMemory("r0", "-", "1")),
			},
			expected: BaseInstruction{
				OpType: Control,
				RMem:   0,
				Mode:   0b000,
				Flag:   0b0100,
				Imm:    -1,
			}},
	}

	runTests(t, &tests)
}
