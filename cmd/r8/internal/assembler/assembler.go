package assembler

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/leon332157/risc-y-8/cmd/r8/internal/assembler/grammar"
	. "github.com/leon332157/risc-y-8/pkg/types"
)

// parse 16 bit two's complement immediate value
func parseImm(imm string) (int16, error) {
	var i64 int64
	//var u64 uint64
	var ret int16
	var err error

	i64, err = strconv.ParseInt(imm, 0, 17) // attempt to parse as signed
	if err != nil {
		return 0, err
	}
	if i64 < -65536 {
		return 0, fmt.Errorf("[parseImm] immediate value less than -65536 : %s", imm)
	}
	if i64 > 65535 {
		return 0, fmt.Errorf("[parseImm] immediate value greater than 65535 : %s", imm)
	}
	// if good then return signed value
	ret = int16(i64)
	//return int16(u64), nil
	return ret, nil
}

// Parse a memory operand and return the base register and displacement as signed 16 bit integer
func parseMemory(mem grammar.OperandMemory) (uint8, int16, error) {
	var rmem uint8 = 0
	var disp int16 = 0

	var err error
	var ok bool
	if mem.Value.Base == "pc" {
		mem.Value.Base = "r0" // pc is encoded as r0
	}
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
		err = fmt.Errorf("[parseMemory] invalid displacement: %v", err)
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
			OpType:   Control,
			RMem:     0x00,
			CtrlFlag: Conditions["unc"].Flag,
			CtrlMode:  Conditions["unc"].Mode,
			Imm:      -1,
		}
	case "ret":
		// encoded as "bunc [lr]"
		ret = BaseInstruction{
			OpType:   Control,
			RMem:     IntegerRegisters["lr"],
			CtrlFlag: Conditions["unc"].Flag,
			CtrlMode:  Conditions["unc"].Mode,
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
			OpType:  LoadStore,
			Rd:      rd,
			MemMode: PUSH,
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
			OpType:  LoadStore,
			Rd:      rd,
			MemMode: POP,
		}
	case "call":
		mem, ok := inst.Operands[0].(grammar.OperandMemory)
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
			return ret, err
		}
		rmem, disp, err := parseMemory(mem)
		if err != nil {
			err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %v %s", mem.Value, err)
			return ret, err
		}
		ret = BaseInstruction{
			OpType:   Control,
			RMem:     rmem,
			CtrlFlag: Conditions["call"].Flag,
			CtrlMode:  Conditions["call"].Mode,
			Imm:      disp,
		}
	case "bunc", "beq", "bz", "bne", "bnz", "blt", "bge", "blu", "bae", "ba", "bof", "bnf":
		mem, ok := inst.Operands[0].(grammar.OperandMemory)
		if !ok {
			err = fmt.Errorf("[parseInstOneOp] invalid operand type: %v", reflect.TypeOf(inst.Operands[0]))
			return ret, err
		}
		rmem, disp, err := parseMemory(mem)
		if err != nil {
			err = fmt.Errorf("[parseInstOneOp] invalid memory operand: %+v %s", mem.Value, err)
			return ret, err
		}
		if cond, ok := Conditions[inst.Mnemonic[1:]]; ok {
			ret = BaseInstruction{
				OpType:   Control,
				RMem:     rmem,
				CtrlFlag: cond.Flag,
				CtrlMode:  cond.Mode,
				Imm:      disp,
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

	switch inst.Mnemonic {
	case "ldw":
		ret = BaseInstruction{
			OpType:  LoadStore,
			Rd:      rd,
			MemMode: LDW,
			RMem:    rmem,
			Imm:     disp,
		}
	case "stw":
		ret = BaseInstruction{
			OpType:  LoadStore,
			Rd:      rd,
			MemMode: STW,
			RMem:    rmem,
			Imm:     disp,
		}
	default:
		err = fmt.Errorf("[parseRMem] invalid instruction: %s", inst.Mnemonic)
		return ret, err
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

var Instructions []BaseInstruction
var Labels map[string]uint32

// func parseLabel(label *grammar.Label) (uint32, error) {

// 	var err error
// 	Labels[label.Text] = label.Offset
// 	return label.Offset, err
	
// }

func ParseLines(lines []grammar.Line) (*[]BaseInstruction,error) {
	for _, line := range lines {
		if line.Directive != nil {
			// TODO: handle directives
			continue
		}
		if line.Label != nil {
			// TODO: handle labels
			Labels[line.Label.Text] = line.Label.Offset
			continue
		}
		if line.Instruction != nil {
			inst, err := parseInst(line.Instruction)
			if err != nil {
				return nil,fmt.Errorf("[parseLines] invalid instruction at position %v: %v", line.Pos, err)
			}
			Instructions = append(Instructions, inst)
		}
	}
	return &Instructions,nil
}

func EncInstructions(insts *[]BaseInstruction) []uint32 {
	enc := make([]uint32, len(*insts))
	for i, inst := range *insts {
		enc[i] = inst.Encode()
	}
	return enc
}
