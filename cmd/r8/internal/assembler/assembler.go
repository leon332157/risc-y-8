package assembler

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/leon332157/risc-y-8/cmd/r8/internal/assembler/grammar"
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

var (
	// ControlOp
	NE   = ControlOp{Mode: 0b000, Flag: 0b0000}
	NZ   = ControlOp{Mode: 0b000, Flag: 0b0000}
	EQ   = ControlOp{Mode: 0b000, Flag: 0b0001}
	Z    = ControlOp{Mode: 0b000, Flag: 0b0001}
	LT   = ControlOp{Mode: 0b001, Flag: 0b0110}
	GE   = ControlOp{Mode: 0b011, Flag: 0b0110}
	LU   = ControlOp{Mode: 0b100, Flag: 0b1000}
	AE   = ControlOp{Mode: 0b000, Flag: 0b1000}
	A    = ControlOp{Mode: 0b010, Flag: 0b1000}
	OF   = ControlOp{Mode: 0b100, Flag: 0b0100}
	NF   = ControlOp{Mode: 0b000, Flag: 0b0100}
	UNC  = ControlOp{Mode: 0b111, Flag: 0b0000}
	CALL = ControlOp{Mode: 0b111, Flag: 0b1111}
)

var Conditions = map[string]ControlOp{
	"ne":   NE,
	"nz":   NZ,
	"eq":   EQ,
	"z":    Z,
	"lt":   LT,
	"ge":   GE,
	"lu":   LU,
	"ae":   AE,
	"a":    A,
	"of":   OF,
	"nf":   NF,
	"unc":  UNC,
	"call": CALL,
}

const (
	IMM_ADD = iota
	IMM_SUB
	IMM_MUL
	IMM_AND
	IMM_XOR
	IMM_OR
	IMM_NOT
	IMM_NEG
	IMM_SHR
	IMM_SAR
	IMM_SHL
	IMM_ROL
	IMM_LDI
	IMM_LDX
	IMM_CMP
)

var ImmALU = map[string]uint8{
	"add": IMM_ADD,
	"sub": IMM_SUB,
	"mul": IMM_MUL,
	"and": IMM_AND,
	"xor": IMM_XOR,
	"orr": IMM_OR,
	"or":  IMM_OR,
	"not": IMM_NOT,
	"neg": IMM_NEG,
	"shr": IMM_SHR,
	"sar": IMM_SAR,
	"shl": IMM_SHL,
	"rol": IMM_ROL,
	"ror": IMM_ROL,
	"ldi": IMM_LDI,
	"ldx": IMM_LDX,
	"cmp": IMM_CMP,
}

const (
	REG_ADD = iota
	REG_SUB
	REG_MUL
	REG_DIV
	REG_REM
	REG_OR
	REG_XOR
	REG_AND
	REG_NOT
	REG_SHL
	REG_SHR
	REG_SAR
	REG_ROL
	REG_CMP
	REG_CPY
	REG_MOV
	REG_NSA
)

var RegALU = map[string]uint8{
	"add": REG_ADD,
	"sub": REG_SUB,
	"mul": REG_MUL,
	"div": REG_DIV,
	"rem": REG_REM,
	"orr": REG_OR,
	"or":  REG_OR,
	"xor": REG_XOR,
	"and": REG_AND,
	"not": REG_NOT,
	"shl": REG_SHL,
	"shr": REG_SHR,
	"sar": REG_SAR,
	"rol": REG_ROL,
	"cmp": REG_CMP,
	"cpy": REG_CPY,
	"mov": REG_MOV,
	"nsa": REG_NSA,
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

const (
	RegImm    = 0b00 // reg-imm
	RegReg    = 0b01 // reg-reg
	LoadStore = 0b10 // load/store
	Control   = 0b11 // control
)

type BaseInstruction struct {
	OpType uint8 // 00 for reg-imm, 01 for reg-reg, 10 for load/store, 11 for control
	Rd     uint8 // Destination register
	ALU    uint8 // ALU operation
	Rs     uint8 // Source register
	RMem   uint8 // Memory register
	Flag   uint8 // Flag for control instructions
	Mode   uint8 // Mode for load/store instructions (0 for load, 10 pop, 10 for store, 11 push)
	Imm    int16 // 16 bit twos complement immediate value
}

// parse 16 bit two's complement immediate value
func parseImm(imm string) (int16, error) {
	temp, err := strconv.ParseInt(imm, 0, 16)
	if err != nil {
		return 0, err
	}
	return int16(temp), err
}

// Parse a memory operand and return the base register and displacement as signed 16 bit integer
func parseMemory(mem grammar.OperandMemory) (uint8, int16, error) {
	var rmem uint8 = 0
	var disp int16 = 0

	var err error
	var ok bool
	rmem, ok = IntegerRegisters[mem.Value.Base] // register operand
	if !ok {
		err = fmt.Errorf("[parseMemory] invalid base register: %#v", mem.Value.Base)
		return 0, 0, err
	}
	// Handle empty operation
	if mem.Value.Operation == "" {
		mem.Value.Operation = "+"
	}
	// handle empty displacement, set to 0
	if mem.Value.Displacement.Value == "" {
		mem.Value.Displacement.Value = "0"
	}
	// check to make sure that not both operation and displacement are negative
	if mem.Value.Displacement.Value[0] == '-' && mem.Value.Operation == "-" {
		err = fmt.Errorf("[parseMemory] invalid displacement, both operation and displacement are negative: %+v", mem.Value)
		return 0, 0, err
	}
	// add the sign to the displacement
	if mem.Value.Operation == "-" {
		mem.Value.Displacement.Value = "-" + mem.Value.Displacement.Value
	}
	// parse the displacement
	disp, err = parseImm(mem.Value.Displacement.Value)
	if err != nil {
		err = fmt.Errorf("[parseMemory] invalid displacement: parseDisp: %s %v", mem.Value.Displacement.Value, err)
		return 0, 0, err
	}
	return rmem, disp, err
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
	case "hlt", "meow":
		// encoded as "bunc [r0+0xFFFF]"
		ret = BaseInstruction{
			OpType: Control,
			RMem:   0x00,
			Flag:   Conditions["unc"].Flag,
			Mode:   Conditions["unc"].Mode,
			Imm:    -1,
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
	case "not", "neg":
		rdval, ok := inst.Operands[0].(grammar.OperandRegister)
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
			return ret, err
		}
		rd, ok := IntegerRegisters[rdval.Value] // register operand
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid register: %#v", rdval.Value)
			return ret, err
		}
		ret = BaseInstruction{
			OpType: RegImm,
			Rd:     rd,
			ALU:    ImmALU[inst.Mnemonic],
		}
	case "push":
		rdval, ok := inst.Operands[0].(grammar.OperandRegister)
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
			return ret, err
		}
		rd, ok := IntegerRegisters[rdval.Value] // register operand
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid register: %#v", rdval.Value)
			return ret, err
		}
		ret = BaseInstruction{
			OpType: LoadStore,
			Rd:     rd,
			Mode:   PUSH,
		}
	case "pop":
		rdval, ok := inst.Operands[0].(grammar.OperandRegister)
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
			return ret, err
		}
		rd, ok := IntegerRegisters[rdval.Value] // register operand
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid register: %#v", rdval.Value)
			return ret, err
		}
		ret = BaseInstruction{
			OpType: LoadStore,
			Rd:     rd,
			Mode:   POP,
		}
	case "call":
		mem, ok := inst.Operands[0].(grammar.OperandMemory)
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
			return ret, err
		}
		rmem, disp, err := parseMemory(mem)
		if err != nil {
			err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
			return ret, err
		}
		ret = BaseInstruction{
			OpType: Control,
			RMem:   rmem,
			Flag:   Conditions["call"].Flag,
			Mode:   Conditions["call"].Mode,
			Imm:    disp,
		}
	case "bunc", "beq", "bz", "bne", "bnz", "blt", "bge", "blu", "bae", "ba", "bof", "bnf":
		mem, ok := inst.Operands[0].(grammar.OperandMemory)
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
			return ret, err
		}
		rmem, disp, err := parseMemory(mem)
		if err != nil {
			err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
			return ret, err
		}
		if cond, ok := Conditions[inst.Mnemonic[1:]]; ok {
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   cond.Flag,
				Mode:   cond.Mode,
				Imm:    disp,
			}
		} else {
			err = fmt.Errorf("[parseInstOneOp] invalid condition code %v on instruction %v", inst.Mnemonic[1:], inst.Mnemonic)
			return ret, err
		}
	/*
		case "beq", "bz":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["eq"].Flag,
				Mode:   Conditions["eq"].Mode,
				Imm:    disp,
			}
		case "bne", "bnz":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["ue"].Flag,
				Mode:   Conditions["ue"].Mode,
				Imm:    disp,
			}
		case "blt":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["lt"].Flag,
				Mode:   Conditions["lt"].Mode,
				Imm:    disp,
			}
		case "bge":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["ge"].Flag,
				Mode:   Conditions["ge"].Mode,
				Imm:    disp,
			}
		case "blu":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["lu"].Flag,
				Mode:   Conditions["lu"].Mode,
				Imm:    disp,
			}
		case "bae":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["ae"].Flag,
				Mode:   Conditions["ae"].Mode,
				Imm:    disp,
			}
		case "ba":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["a"].Flag,
				Mode:   Conditions["a"].Mode,
				Imm:    disp,
			}
		case "bof":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret = BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["of"].Flag,
				Mode:   Conditions["of"].Mode,
				Imm:    disp,
			}
		case "bnf":
			mem, ok := inst.Operands[0].(grammar.OperandMemory)
			if !ok {
				err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
				return ret, err
			}
			rmem, disp, err := parseMemory(mem)
			if err != nil {
				err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v err: %s", mem.Value, err)
				return ret, err
			}
			ret =  BaseInstruction{
				OpType: Control,
				RMem:   rmem,
				Flag:   Conditions["nf"].Flag,
				Mode:   Conditions["nf"].Mode,
				Imm:    disp,
			}
	*/
	default:
		err = fmt.Errorf("[parseInstOneOp] invalid instruction: %s at %v", inst.Mnemonic, inst.Pos)
		return ret, err
	}
	return ret, nil
}

// Parse instructions with two opernads
func parseInstTwoOp(inst *grammar.Instruction) (BaseInstruction, error) {
	var ret BaseInstruction
	var err error

	rdval, ok := inst.Operands[0].(grammar.OperandRegister)
	if !ok {
		err = fmt.Errorf("[parseInstTwoOp] invalid operand 1 type: %T", inst.Operands[0])
		return ret, err
	}
	rd, ok := IntegerRegisters[rdval.Value] // register operand
	if !ok {
		err = fmt.Errorf("[parseInstTwoOp] invalid destination register: %#v", rdval.Value)
		return ret, err
	}
	switch inst.Operands[1].(type) {
	case grammar.OperandImmediate:
		return parseRI(inst, rd)
	case grammar.OperandRegister:
		return parseRR(inst, rd)
	case grammar.OperandMemory:
		return parseRMem(inst, rd)
	default:
		err = fmt.Errorf("[parseInstTwoOp] invalid operand 2 type: %T", inst.Operands[1])
	}
	return ret, err
}

// Parse register and immediate instructions
func parseRI(inst *grammar.Instruction, rd uint8) (BaseInstruction, error) {
	var ret BaseInstruction
	var err error

	alu, ok := ImmALU[inst.Mnemonic]
	if !ok {
		err = fmt.Errorf("[parserRI] invalid ALU operation %v", inst.Mnemonic)
		return ret, err
	}
	imm, err := parseImm(inst.Operands[1].(grammar.OperandImmediate).Value)
	if err != nil {
		return ret, err
	}

	switch inst.Mnemonic {
	case "add", "sub", "mul", "and", "xor", "or", "orr":
		break
	case "shr", "sar", "shl":
		if imm < 0 {
			err = fmt.Errorf("[parserRI] invalid negative immediate value for shift: %v", imm)
			return ret, err
		}
		if imm > 31 {
			err = fmt.Errorf("[parseRI] immediate value for shift is greater than 31: %v", imm)
			return ret, err
		}
	case "rol":
		if imm < 0 {
			err = fmt.Errorf("[parserRI] invalid negative immediate value for rotate right: %v", imm)
			return ret, err
		}
		if imm > 31 {
			err = fmt.Errorf("[parserRI] immediate value for rotate left is greater than 31: %v", imm)
			return ret, err
		}
	case "ror":
		if imm < 0 {
			err = fmt.Errorf("[parserRI] invalid negative immediate value for rotate right: %v", imm)
			return ret, err
		}
		if imm > 31 {
			err = fmt.Errorf("[parserRI] immediate value for rotate right is greater than 31: %v", imm)
			return ret, err
		}
		imm = (32 - imm) // rotate right is 32 - imm
	case "ldi", "ldx", "cmp":
		break
	}
	ret = BaseInstruction{
		OpType: RegImm,
		Rd:     rd,
		ALU:    alu,
		Imm:    imm,
	}
	return ret, err
}

// Parse register to register instructions
func parseRR(inst *grammar.Instruction, rd uint8) (BaseInstruction, error) {
	var ret BaseInstruction
	var err error

	alu, ok := RegALU[inst.Mnemonic]
	if !ok {
		err = fmt.Errorf("[parserRR] invalid ALU operation %v", inst.Mnemonic)
		return ret, err
	}

	rsval, ok := inst.Operands[1].(grammar.OperandRegister)
	if !ok {
		err = fmt.Errorf("[parserRR] invalid operand type: %T", inst.Operands[1])
		return ret, err
	}
	rs, ok := IntegerRegisters[rsval.Value] // register operand
	if !ok {
		err = fmt.Errorf("[parserRR] invalid source register: %#v", rsval.Value)
		return ret, err
	}

	switch inst.Mnemonic {
	case "add", "sub", "mul", "div", "rem", "and", "xor", "or", "orr":
		break
	case "shr", "sar", "shl", "rol":
		break
	case "ror":
		break
	case "cmp", "cpy", "nsa":
		break
	default:
		err = fmt.Errorf("[parserRR] invalid instruction: %s", inst.Mnemonic)
	}
	ret = BaseInstruction{
		OpType: RegReg,
		Rd:     rd,
		ALU:    alu,
		Rs:     rs,
	}
	return ret, err
}

// Parse memory to register instructions
func parseRMem(inst *grammar.Instruction, rd uint8) (BaseInstruction, error) {
	var ret BaseInstruction
	var err error

	memval, ok := inst.Operands[1].(grammar.OperandMemory)
	if !ok {
		err = fmt.Errorf("[parseRMem] invalid operand type: %T", inst.Operands[1])
		return ret, err
	}
	rmem, disp, err := parseMemory(memval)
	if err != nil {
		err = fmt.Errorf("[parseRMem] invalid memory operand: %+v err: %s", memval.Value, err)
		return ret, err
	}
	alu, ok := RegALU[inst.Mnemonic]
	if !ok {
		err = fmt.Errorf("[parseRMem] invalid ALU operation %v", inst.Mnemonic)
		return ret, err
	}

	switch inst.Mnemonic {
	case "ldw", "stw":
		break
	default:
		err = fmt.Errorf("[parseRMem] invalid instruction: %s", inst.Mnemonic)
		return ret, err
	}
	ret = BaseInstruction{
		OpType: LoadStore,
		Rd:     rd,
		RMem:   rmem,
		ALU:    alu,
		Imm:    disp,
	}
	return ret, err
}

func parseInst(inst *grammar.Instruction) (BaseInstruction, error) {
	// Parse the instruction based on the grammar rules, and return a slice of BaseInstruction if pseudo instructions are found.
	//var instSlice = make([]BaseInstruction, 0, 2)
	switch len(inst.Operands) {
	case 0:
		// no operands
		return parseInstNoOp(inst)
	case 1:
		// one operand
		return parseInstOneOp(inst)
	case 2:
		// two operands
		return parseInstTwoOp(inst)
	default:
		err := fmt.Errorf("[parseInst] invalid number of operands: %d", len(inst.Operands))
		return BaseInstruction{}, err
	}
}
