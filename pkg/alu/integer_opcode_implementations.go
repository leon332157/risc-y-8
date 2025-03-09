package alu
import (
	"math/bits"
	"github.com/leon332157/risc-y-8/pkg/types"
)

type CPU struct {

	// Special Registers
	ZeroRegister 	uint32
	RFlag 			uint32
	VFlag			uint32
	VType			uint32
	PC 				uint32

	// General Purpose Registers
	IntRegisters 	[16]uint32

	// Floating Point Registers
	FPRegisters 	[8]float32

	// Vector Registers
	VecRegisters	[8]uint32
}

func (cpu *CPU) SetFlag(flag uint32) {
	cpu.RFlag |= flag
}

func (cpu *CPU) ClearFlag(flag uint32) {
	cpu.RFlag &^= flag
}

func (cpu *CPU) CheckFlag(flag uint32) bool {
	return (cpu.RFlag & flag) != 0
}

func (cpu *CPU) ResetFlags() {
	cpu.RFlag = 0
} 

func SignExtend(val uint32) int64 {
	return int64(int32(val))
}

/* Adds an immediate (source) to rd. Calculation is signed arithmetic. CF is set when carry
out is 1. ZF is set when the addition result is 0, SF is set when the result of addition is
negative, OF is calculated as SF ^ MSB of imm ^ MSB of rd */
func (cpu *CPU) ADD(rd uint8, imm int32) {

	// If rd is 0, do nothing
	if rd == 0x00 {
		return
	}

	
	a := cpu.IntRegisters[rd-1]
	b := uint32(imm)

	result := a + b

	// Reset previous flags
	cpu.ResetFlags()

	if result < a {
		cpu.SetFlag(types.CF)
	}

	if result == 0 {
		cpu.SetFlag(types.ZF)
	}

	if result & 0x80000000 != 0 {
		cpu.SetFlag(types.SF)
	}

	if ((a^b) & 0x80000000 == 0) && ((a^result) & 0x80000000 != 0) {
		cpu.SetFlag(types.OF)
	}

	cpu.IntRegisters[rd-1] = result

}

// rd = imm - rd, performed as rd + (-1*imm) Flag behavior follows ADD
func (cpu *CPU) SUB(rd uint8, imm int32) {

	// If rd is 0, do nothing
	if rd == 0x00 {
		return
	}

	a := cpu.IntRegisters[rd-1]
	b := uint32(imm)

	result := b + (^a + 1)

	// Reset previous flags
	cpu.ResetFlags()

	if result == 0 {
		cpu.SetFlag(types.ZF)
	}

	if result & 0x80000000 != 0 {
		cpu.SetFlag(types.SF)
	}

	if ((b^a) & 0x80000000 != 0) && ((b^result) & 0x80000000 != 0) {
		cpu.SetFlag(types.OF)
	}

	cpu.IntRegisters[rd-1] = result

}

func (cpu *CPU) MUL(rd uint8, imm int32) {

	// If rd is 0, do nothing
	if rd == 0x00 {
		return
	}

	a := cpu.IntRegisters[rd-1]
	b := uint32(imm)

	hi, lo := bits.Mul32(a, b)

	// Reset previous flags
	cpu.ResetFlags()

	if (hi | lo) == 0 {
		cpu.SetFlag(types.ZF)
	}

	if hi != 0 {
		cpu.SetFlag(types.CF)
	}

	if SignExtend(lo) != (int64(hi) << 32 | int64(lo)) {
		cpu.SetFlag(types.OF)
	}

	cpu.IntRegisters[rd-1] = lo

}