
package alu
import (
	"math/bits"
	"github.com/leon332157/risc-y-8/pkg/types"
)

func SignExtend(val uint32) int64 {
	return int64(int32(val))
}

/* Adds an immediate (source) to rd. Calculation is signed arithmetic. CF is set when carry out is 1. ZF is set when the addition result is 0, SF is set when the result of addition is negative, OF is calculated as SF ^ MSB of imm ^ MSB of rd */
func (reg *Registers) IMM_ADD(rd uint8, imm int16) types.MemoryStageInput{

	msi := types.MemoryStageInput{}

	// If rd is 0, do nothing
	// if rd == 0x00 {
	// 	return
	// }
	
	augend := reg.IntRegisters[rd-1]
	addend := uint32(imm)

	result, carry := bits.Add32(augend, addend, 0)

	msi.Reg = rd
	msi.RegVal = result
	msi.IsALU = true

	// Reset previous flags
	reg.ResetFlags()

	if carry == 1 {
		msi.Flag |= types.CF
	}

	if result == 0 {
		msi.Flag |= types.ZF
	}

	if result & 0x80000000 != 0 {
		msi.Flag |= types.SF
	}

	if ((augend ^ addend) & 0x80000000 == 0) && ((augend ^ result) & 0x80000000 != 0) {
		msi.Flag |= types.OVF
	}

	return msi

}

// rd = imm - rd, performed as rd + (-1*imm) Flag behavior follows ADD
func (regs *Registers) IMM_SUB(rd uint8, imm int16) types.MemoryStageInput{

	msi := types.MemoryStageInput{}

	// if rd == 0x00 {
	// 	return
	// }

	subtrahend := uint32(imm)
	minuend := regs.IntRegisters[rd-1]

	result, carry := bits.Sub32(subtrahend, minuend, 0)

	msi.Reg = rd
	msi.RegVal = result
	msi.IsALU = true

	regs.ResetFlags()
	if carry == 1 {
		msi.Flag |= types.CF
	}

	if result == 0 {
		msi.Flag |= types.ZF
	}

	if result&0x80000000 != 0 {
		msi.Flag |= types.SF
	}

	if ((subtrahend ^ minuend) & 0x80000000 != 0) && ((subtrahend ^ result)&0x80000000 != 0) {
		msi.Flag |= types.OVF
	}

	return msi

}

func (regs *Registers) REG_MUL(rd uint8, rs uint8) types.MemoryStageInput{

	msi := types.MemoryStageInput{}

	// if rd == 0x00 {
	// 	return
	// }

	multiplicand := regs.IntRegisters[rd-1]
	multiplier := regs.IntRegisters[rs-1]

	hi, lo := bits.Mul32(multiplicand, multiplier)

	msi.Reg = rd
	msi.RegVal = lo

	regs.ResetFlags()

	if (hi | lo) == 0 {
		msi.Flag |= types.ZF
	}

	if hi != 0 {
		msi.Flag |= types.CF
	}

	if SignExtend(lo) != (int64(hi)<<32 | int64(lo)) {
		msi.Flag |= types.OVF
	}

	return msi

}

func (regs *Registers) IMM_LDW(rd uint8, rs uint8, imm int16) types.MemoryStageInput {

	return types.MemoryStageInput{
		WriteBackStageInput: types.WriteBackStageInput{
			Reg:	rd,
		},
		Address:	int(regs.IntRegisters[rs-1]) + int(imm), 
		IsLoad:		true,
	}

}

func (regs *Registers) IMM_STW(rd uint8, rs uint8, imm int16) types.MemoryStageInput {

	return types.MemoryStageInput{
		Address:	int(regs.IntRegisters[rs-1]) + int(imm), 
		Data:		regs.IntRegisters[rd-1],
		IsLoad:		false,
	}

}

// exact same as subtract but doesnt set the destination register.
func (regs *Registers) IMM_CMP(rd uint8, imm int16) types.MemoryStageInput{

	msi := types.MemoryStageInput{}

	// if rd == 0x00 {
	// 	return
	// }

	subtrahend := uint32(imm)
	minuend := regs.IntRegisters[rd-1]

	result, carry := bits.Sub32(subtrahend, minuend, 0)

	regs.ResetFlags()
	if carry == 1 {
		msi.Flag |= types.CF
	}

	if result == 0 {
		msi.Flag |= types.ZF
	}

	if result&0x80000000 != 0 {
		msi.Flag |= types.SF
	}

	if ((subtrahend ^ minuend) & 0x80000000 != 0) && ((subtrahend ^ result)&0x80000000 != 0) {
		msi.Flag |= types.OVF
	}

	return msi

}

// JMP if ZF == 0
func (regs *Registers) BNE(rmem uint8, imm int16) types.MemoryStageInput{

	msi := types.MemoryStageInput{}
	msi.IsControl = true

	if !regs.CheckFlag(types.ZF) {
		newAddress := int(uint16(rmem) + uint16(imm))
		msi.Branch_PC = uint32(newAddress)
	}

	return msi

}