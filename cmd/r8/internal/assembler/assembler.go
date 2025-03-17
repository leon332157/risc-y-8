package assembler

import (
	_ "errors"
	"fmt"
	"github.com/leon332157/risc-y-8/cmd/r8/internal/assembler/grammar"
	_ "strings"
)

var IntegerRegisters = map[string]uint8{
	"r0": 0x00, // r0 is reserved for the zero register
	"r1": 0x01, "r2": 0x02, "r3": 0x03, "r4": 0x04,
	"r5": 0x05, "r6": 0x06, "r7": 0x07, "r8": 0x08,
	"r9": 0x09, "r10": 0x0A, "r11": 0x0B, "r12": 0x0C,
	"r13": 0x0D, "r14": 0x0E, "r15": 0x0F, "r16": 0x10,
	"r17": 0x11, "r18": 0x12, "r19": 0x13, "r20": 0x14,
	"r21": 0x15, "r22": 0x16, "r23": 0x17, "r24": 0x18,
	"r25": 0x19, "r26": 0x1A, "r27": 0x1B, "r28": 0x1C,
	"r29": 0x1D, "r30": 0x1E, "r31": 0x1F,
	"bp": 29, // base pointer
	"sp": 30, // stack pointer
	"lr": 31, // link register
}

var FPRegisters = map[string]uint8{
	"f1": 0x00, "f2": 0x01, "f3": 0x02, "f4": 0x03,
	"f5": 0x04, "f6": 0x05, "f7": 0x06, "f8": 0x07,
}

var VectorRegisters = map[string]uint8{
	"v1": 0x00, "v2": 0x01, "v3": 0x02, "v4": 0x03,
	"v5": 0x04, "v6": 0x05, "v7": 0x06, "v8": 0x07,
}

// Control
type ControlOp struct {
	Mode uint8 // 3 bits
	Flag uint8 // 4 bits
}

var Conditions = map[string]ControlOp{
	"eq":   {Mode: 0b000, Flag: 0b0000},
	"ne":   {Mode: 0b000, Flag: 0b0001},
	"lt":   {Mode: 0b001, Flag: 0b0110},
	"ge":   {Mode: 0b011, Flag: 0b0110},
	"lu":   {Mode: 0b100, Flag: 0b1000},
	"ae":   {Mode: 0b000, Flag: 0b1000},
	"a":    {Mode: 0b010, Flag: 0b1000},
	"of":   {Mode: 0b100, Flag: 0b0100},
	"nf":   {Mode: 0b000, Flag: 0b0100},
	"unc":  {Mode: 0b111, Flag: 0b0000},
	"call": {Mode: 0b000, Flag: 0b1111},
}

var ImmALU = map[string]uint8{
	"add": 0b0000,
	"sub": 0b0001,
	"mul": 0b0010,
	"and": 0b0011,
	"xor": 0b0100,
	"orr": 0b0101,
	"not": 0b0110,
	"neg": 0b0111,
	"shr": 0b1000,
	"sar": 0b1001,
	"shl": 0b1010,
	"rol": 0b1011,
	"ldi": 0b1100,
	"ldx": 0b1101,
	"cmp": 0b1110,
}

var RegALU = map[string]uint8{
	"add": 0b0000,
	"sub": 0b0001,
	"mul": 0b0010,
	"div": 0b0011,
	"rem": 0b0100,
	"orr": 0b0101,
	"xor": 0b0110,
	"and": 0b0111,
	"not": 0b1000,
	"shl": 0b1001,
	"shr": 0b1010,
	"sar": 0b1011,
	"rol": 0b1100,
	"cmp": 0b1101,
	"cpy": 0b1110,
}

const (
	LDW  = iota // Load Word
	POP         // Pop
	PUSH        // Load X
	STW         // Store Word
)

type DataType uint8

const (
	//reserved	DataType = 0b00
	Integer DataType = 0b01
	Float   DataType = 0b10
	Vector  DataType = 0b11
)

type Instruction interface {
	Encode() uint32
}

const (
	RegImm    = 0b00 // reg-imm
	RegReg    = 0b01 // reg-reg
	LoadStore = 0b10 // load/store
	Control   = 0b11 // control
)

type BaseInstruction struct {
	OpType uint8  // 00 for reg-imm, 01 for reg-reg, 10 for load/store, 11 for control
	Rd     uint8  // Destination register
	ALU    uint8  // ALU operation
	Rs     uint8  // Source register
	RMem   uint8  // Memory register
	Flag   uint8  // Flag for control instructions
	Mode   uint8  // Mode for load/store instructions (0 for load, 1 for store, 01 pop, 10 push)
	Imm    uint16 // 16 bit immediate value
}

func parseInstNoOp(inst *grammar.Instruction) (BaseInstruction, error) {
	var ret BaseInstruction
	var err error

	switch inst.Mnemonic {
	case "nop":
		// encoded as "cpy r0, r0"
		ret = BaseInstruction{
			OpType: RegReg,
			Rd:     0x00, // r0
			ALU:    RegALU["cpy"],
			Rs:     0x00, // r0
		}
	case "hlt":
		// encoded as "bunc [r0+0xFFFF]"
		ret = BaseInstruction{
			OpType: Control,
			RMem:   0x00,
			Flag:   Conditions["unc"].Flag,
			Mode:   Conditions["unc"].Mode,
			Imm:    0xffff,
		}
	case "ret":
		// encoded as "bunc [lr]"
		ret = BaseInstruction{
			OpType: Control,
			RMem:   IntegerRegisters["lr"],
			Flag:   Conditions["unc"].Flag,
			Mode:   Conditions["unc"].Mode,
		}

	}
	return ret, err
}

func parseInstOneOp(inst *grammar.Instruction) (BaseInstruction, error) {
	var ret BaseInstruction
	var err error

	switch inst.Mnemonic {
	case "push":
		rd, ok := IntegerRegisters[inst.Operands[0].Value()]
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid register: %s", inst.Operands[0].Value())
			return ret, err
		}
		ret = BaseInstruction{
			OpType: LoadStore,
			Rd:     rd,
			Mode:   PUSH,
		}
	case "pop":
		rd, ok := IntegerRegisters[inst.Operands[0].Value()]
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid register: %s", inst.Operands[0].Value())
		}
		goto parseInstOneOpError
		ret = BaseInstruction{
			OpType: LoadStore,
			Rd:     rd,
			Mode:   POP,
		}
	case "call":
		rmem, ok := IntegerRegisters[inst.Operands[0].Value()]
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid register: %s", inst.Operands[0].Value())
			return ret, err
		}
		ret = BaseInstruction{
			OpType: Control,
			RMem:   IntegerRegisters[inst.Operands[0].Value],
			Flag:   Conditions["call"].Flag,
			Mode:   Conditions["call"].Mode,
		}
	case "bunc":
		ret = BaseInstruction{
			OpType: Control,
			RMem:   IntegerRegisters[inst.Operands[0].Value],
			Flag:   Conditions["unc"].Flag,
			Mode:   Conditions["unc"].Mode,
		}
	case "beq":
		ret = BaseInstruction{
			OpType: Control,
			RMem:   IntegerRegisters[inst.Operands[0].Value],
			Flag:   Conditions["eq"].Flag,
			Mode:   Conditions["eq"].Mode,
		}
	case "bne":
		ret = BaseInstruction{
			OpType: Control,
			RMem:   IntegerRegisters[inst.Operands[0].Value()],
			Flag:   Conditions["ne"].Flag,
			Mode:   Conditions["ne"].Mode,
		}
	case "blt":
		ret = BaseInstruction{
			OpType: Control,
			RMem:   IntegerRegisters[inst.Operands[0].Value()],
			Flag:   Conditions["lt"].Flag,
			Mode:   Conditions["lt"].Mode,
		}
	}
parseOneOpError:
	if err != nil {
		return ret, err
	}
return ret, nil
}

func parseOneInst(inst *grammar.Instruction) (BaseInstruction, error) {
	// Parse the instruction based on the grammar rules, and return a slice of BaseInstruction if pseudo instructions are found.
	//var instSlice = make([]BaseInstruction, 0, 2)
	var ret BaseInstruction
	var err error

	switch len(inst.Operands) {
	case 0:
		// no operands
		return parseInstNoOp(inst)
	case 1:
		// one operand
		return parseInstOneOp(inst)
	}
	return ret, err
}
