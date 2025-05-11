package assembler

import (
	"testing"

	"github.com/leon332157/risc-y-8/cmd/r8/assembler/grammar"
	. "github.com/leon332157/risc-y-8/pkg/types"
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
		if *(inst.(*BaseInstruction)) != testCase.expected {
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
			OpType:  LoadStore,
			Rd:      1,
			MemMode: PUSH,
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
				OpType:   Control,
				RMem:     31,
				CtrlMode: 0b111,
				CtrlFlag: 0b1111,
				Imm:      0x100,
			},
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "call",
				Operands: makeOperands(makeMemory("r31", "+", "-1"))},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     31,
				CtrlMode: 0b111,
				CtrlFlag: 0b1111,
				Imm:      -1,
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
				OpType:   Control,
				RMem:     10,
				CtrlMode: 0b111,
				CtrlFlag: 0b1111,
				Imm:      -0x100,
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
				OpType:   Control,
				RMem:     31,
				CtrlMode: 0b111,
				CtrlFlag: 0b0000,
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
				OpType:   Control,
				RMem:     1,
				CtrlMode: 0b0,
				CtrlFlag: 0b0,
				Imm:      0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bz",
				Operands: makeOperands(
					makeMemory("r9", "+", "-10"))},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     9,
				CtrlMode: 0b0,
				CtrlFlag: 0b1,
				Imm:      -10,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "blt",
				Operands: makeOperands(
					makeMemory("r29", "-", "0x111"))},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     29,
				CtrlMode: 0b001,
				CtrlFlag: 0b0110,
				Imm:      -0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bge",
				Operands: makeOperands(
					makeMemory("r31", "-", "0x111"))},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     31,
				CtrlMode: 0b011,
				CtrlFlag: 0b0110,
				Imm:      -0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "blu",
				Operands: makeOperands(
					makeMemory("r31", "-", "0x111"))},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     31,
				CtrlMode: 0b100,
				CtrlFlag: 0b1000,
				Imm:      -0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bae",
				Operands: makeOperands(
					makeMemory("r31", "-", "0x111"))},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     31,
				CtrlMode: 0b000,
				CtrlFlag: 0b1000,
				Imm:      -0x111,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "ba",
				Operands: makeOperands(
					makeMemory("r31", "", ""))},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     31,
				CtrlMode: 0b010,
				CtrlFlag: 0b1000,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bof",
				Operands: makeOperands(
					makeMemory("r0", "", "")),
			},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     0,
				CtrlMode: 0b100,
				CtrlFlag: 0b0100,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "bnf",
				Operands: makeOperands(
					makeMemory("r0", "-", "1")),
			},
			expected: BaseInstruction{
				OpType:   Control,
				RMem:     0,
				CtrlMode: 0b000,
				CtrlFlag: 0b0100,
				Imm:      -1,
			}},
	}

	runTests(t, &tests)
}

func TestRegImm(t *testing.T) {
	var tests = []testCase{
		{
			instr: grammar.Instruction{
				Mnemonic: "add",
				Operands: makeOperands(
					makeRegister("r1"),
					makeImmediate("0x100"))},
			expected: BaseInstruction{
				OpType: RegImm,
				Rd:     1,
				ALU:    0b0000,
				Imm:    0x100,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "ror",
				Operands: makeOperands(
					makeRegister("r1"),
					makeImmediate("64"),
				)},
			expected:      BaseInstruction{},
			errorExpected: true,
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "ror",
				Operands: makeOperands(
					makeRegister("r1"),
					makeImmediate("11"))},
			expected: BaseInstruction{
				OpType: RegImm,
				Rd:     1,
				ALU:    0b1011,
				Imm:    21,
			},
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "rol",
				Operands: makeOperands(
					makeRegister("r1"),
					makeImmediate("21"))},
			expected: BaseInstruction{
				OpType: RegImm,
				Rd:     1,
				ALU:    0b1011,
				Imm:    10,
			},
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "rol",
				Operands: makeOperands(
					makeRegister("r1"),
					makeImmediate("-1"))},
			expected:      BaseInstruction{},
			errorExpected: true,
		},
		{
			instr: grammar.Instruction{
				Mnemonic: "cmp",
				Operands: makeOperands(
					makeRegister("r31"),
					makeImmediate("-1"))},
			expected: BaseInstruction{
				OpType: RegImm,
				Rd:     31,
				ALU:    0b1110,
				Imm:    -1,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "sub",
				Operands: makeOperands(
					makeRegister("r30"),
					makeImmediate("-65535"))},
			expected:      BaseInstruction{},
			errorExpected: true,
		},
	}

	runTests(t, &tests)
}

func TestRegReg(t *testing.T) {
	var tests = []testCase{
		{
			instr: grammar.Instruction{
				Mnemonic: "add",
				Operands: makeOperands(
					makeRegister("r1"),
					makeRegister("r2"))},
			expected: BaseInstruction{
				OpType: RegReg,
				Rd:     1,
				Rs:     2,
				ALU:    0b0000,
			}},
		{
			instr: grammar.Instruction{
				Mnemonic: "cmp",
				Operands: makeOperands(
					makeRegister("r31"),
					makeRegister("r64"),
				)},
			expected:      BaseInstruction{},
			errorExpected: true,
		},
	}

	runTests(t, &tests)
}
