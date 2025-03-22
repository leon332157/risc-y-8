package assembler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

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

var Conditions = map[string]ControlOp{
	"ne":   {Mode: 0b000, Flag: 0b0000},
	"eq":   {Mode: 0b000, Flag: 0b0001},
	"lt":   {Mode: 0b001, Flag: 0b0110},
	"ge":   {Mode: 0b011, Flag: 0b0110},
	"lu":   {Mode: 0b100, Flag: 0b1000},
	"ae":   {Mode: 0b000, Flag: 0b1000},
	"a":    {Mode: 0b010, Flag: 0b1000},
	"of":   {Mode: 0b100, Flag: 0b0100},
	"nf":   {Mode: 0b000, Flag: 0b0100},
	"unc":  {Mode: 0b111, Flag: 0b0000},
	"call": {Mode: 0b111, Flag: 0b1111},
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

// parse 16 bit unsigned immediate value
func parseImm(imm string) (uint16, error) {
	var ret uint16 = 0
	if strings.HasPrefix(imm, "0x") {
		// hex number
		temp, err := strconv.ParseUint(imm[2:], 16, 16)
		if err != nil {
			return ret, err
		}
		ret = uint16(temp)
	} else if strings.HasPrefix(imm, "0b") {
		// binary number
		temp, err := strconv.ParseUint(imm[2:], 2, 16)
		if err != nil {
			return ret, err
		}
		ret = uint16(temp)
	} else {
		// decimal number
		temp, err := strconv.ParseUint(imm, 10, 16)
		if err != nil {
			return ret, err
		}
		ret = uint16(temp)
	}
	return ret, nil
}

// Parse a 16 bit twos complement displacement value
func parseDisp(disp string) (int16, error) {
	var ret int16 = 0
	// decimal number
	temp, err := strconv.ParseInt(disp, 0, 16)
	if err != nil {
		return ret, err
	}
	ret = int16(temp)
	return ret, nil
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
	disp, err = parseDisp(mem.Value.Displacement.Value)
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
	case "hlt":
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
	case "bunc":
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
			Flag:   Conditions["unc"].Flag,
			Mode:   Conditions["unc"].Mode,
			Imm:    disp,
		}
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
		ret = BaseInstruction{
			OpType: Control,
			RMem:   rmem,
			Flag:   Conditions["nf"].Flag,
			Mode:   Conditions["nf"].Mode,
			Imm:    disp,
		}
	default:
		err = fmt.Errorf("[parseInstOneOp] invalid instruction: %s at %v", inst.Mnemonic, inst.Pos)
		return ret, err
	}
	return ret, nil
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
	}
	return BaseInstruction{}, nil
}
