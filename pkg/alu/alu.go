package alu

import (
	"math/bits"
)

type ALU struct {
	Registers    [31]uint32
	FlagRegister uint32
}

const (
	CF  uint32 = 1 << 0 // Carry Flag
	OVF uint32 = 1 << 1 // Overflow Flag
	ZF  uint32 = 1 << 2 // Zero Flag
	SF  uint32 = 1 << 3 // Sign Flag
)

func SignExtend(val uint32) int64 {
	return int64(int32(val))
}

func (alu *ALU) ResetFlags() {
	alu.FlagRegister = 0
}

func (alu *ALU) SetFlag(flag uint32) {
	alu.FlagRegister |= flag
}

func (alu *ALU) ClearFlag(flag uint32) {
	alu.FlagRegister &= ^flag
}

func (alu *ALU) GetFlag(flag uint32) bool {
	return (alu.FlagRegister & flag) != 0
}

func (alu *ALU) Add(augend, addend uint32) uint32 {

	sum, carry := bits.Add32(augend, addend, 0)

	// Reset previous flags
	alu.ResetFlags()

	if carry == 1 {
		alu.FlagRegister |= CF
	}

	if sum == 0 {
		alu.FlagRegister |= ZF
	}

	if sum&0x80000000 != 0 {
		alu.FlagRegister |= SF
	}

	if ((augend^addend)&0x80000000 == 0) && ((augend^sum)&0x80000000 != 0) {
		alu.FlagRegister |= OVF
	}

	return sum

}

func (alu *ALU) Sub(minuend, subtrahend uint32) uint32 {
	diff, carry := bits.Sub32(minuend, subtrahend, 0)

	// Reset previous flags
	alu.ResetFlags()

	if carry == 1 {
		alu.FlagRegister |= CF
	}

	if diff == 0 {
		alu.FlagRegister |= ZF
	}

	if diff&0x80000000 != 0 {
		alu.FlagRegister |= SF
	}

	if ((minuend^subtrahend)&0x80000000 == 0) && ((minuend^diff)&0x80000000 != 0) {
		alu.FlagRegister |= OVF
	}

	return diff

}

func (alu *ALU) Mul(multiplicand, multiplier uint32) uint32 {
	hi, lo := bits.Mul32(multiplicand, multiplier)

	// Reset previous flags
	alu.ResetFlags()

	if (hi | lo) == 0 {
		alu.FlagRegister |= ZF
	}

	if hi != 0 {
		alu.FlagRegister |= CF
	}

	if SignExtend(lo) != (int64(hi)<<32 | int64(lo)) {
		alu.FlagRegister |= OVF
	}

	return lo

}

func (alu *ALU) Div(dividend, divisor uint32) uint32 {
	// TODO: Implement division
	return 0
}

func (alu *ALU) Rem(dividend, divisor uint32) uint32 {
	// TODO: Implement remainder
	return 0
}

func (alu *ALU) And(a, b uint32) uint32 {
	return a & b
}

func (alu *ALU) Or(a, b uint32) uint32 {
	return a | b
}

func (alu *ALU) Xor(a, b uint32) uint32 {
	return a ^ b
}

func (alu *ALU) Not(a uint32) uint32 {
	return ^a
}

func (alu *ALU) Neg(a uint32) uint32 {
	return -a
}

