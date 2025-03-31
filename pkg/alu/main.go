package alu

import (
	"github.com/leon332157/risc-y-8/pkg/types"
)

var etodecode types.ExeToDecode

func ExecuteStageToMemory(regs *Registers, inst types.DecodeToExe) types.ExeToMem{

	switch inst.Instruction.OpType {

	case types.RegImm:

		switch inst.Instruction.ALU {

		case types.IMM_ADD:
			return regs.IMM_ADD(inst.Instruction.Rd, inst.Instruction.Imm)
		case types.IMM_SUB:
			return regs.IMM_SUB(inst.Instruction.Rd, inst.Instruction.Imm)
		default:
			panic("unsupported immediate ALU operation")
		}

	case types.RegReg:

		switch inst.Instruction.ALU {
		case types.REG_MUL:
			return regs.REG_MUL(inst.Instruction.Rd, inst.Instruction.Rs)
		default:
			panic("unsupported register ALU operation")
		}

	case types.LoadStore:

		switch inst.Instruction.Mode{

		case types.LDW:
			return regs.IMM_STW(inst.Instruction.Rd, inst.Instruction.RMem, inst.Instruction.Imm)
		case types.STW:
			return regs.IMM_LDW(inst.Instruction.Rd, inst.Instruction.RMem, inst.Instruction.Imm)
		default:
			panic("unsupported load/store operation")
		}
	case types.Control:
		
		// BNE
		if (inst.Instruction.Mode == 0b000 && inst.Instruction.Flag == 0b0000) {
			return regs.BNE(inst.Instruction.RMem, inst.Instruction.Imm)
		} else {
			panic("unsupported control operation")
		}
		
	default:
		// Handle unsupported operation type
		panic("unsupported operation type")
	}
}

func ExecuteStageToDecode() types.ExeToDecode {

	return etodecode

}