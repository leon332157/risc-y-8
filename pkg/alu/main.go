package alu

import (
	"github.com/leon332157/risc-y-8/pkg/types"
)

func ExecuteInstruction(regs *Registers, inst types.BaseInstruction) types.MemoryStageInput{

	switch inst.OpType {

	case types.RegImm:

		switch inst.ALU {

		case types.IMM_ADD:
			return regs.IMM_ADD(inst.Rd, inst.Imm)
		case types.IMM_SUB:
			return regs.IMM_SUB(inst.Rd, inst.Imm)
		default:
			panic("unsupported immediate ALU operation")
		}

	case types.RegReg:

		switch inst.ALU {
		case types.REG_MUL:
			return regs.REG_MUL(inst.Rd, inst.Rs)
		default:
			panic("unsupported register ALU operation")
		}

	case types.LoadStore:

		switch inst.Mode{

		case types.LDW:
			return regs.IMM_STW(inst.Rd, inst.RMem, inst.Imm)
		case types.STW:
			return regs.IMM_LDW(inst.Rd, inst.RMem, inst.Imm)
		default:
			panic("unsupported load/store operation")
		}
	case types.Control:
		
		// BNE
		if (inst.Mode == 0b000 && inst.Flag == 0b0000) {
			return regs.BNE(inst.RMem, inst.Imm)
		} else {
			panic("unsupported control operation")
		}
		
	default:
		// Handle unsupported operation type
		panic("unsupported operation type")
	}
}