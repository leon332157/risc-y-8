package alu

import (
	"math/bits"
)

type ALU struct {
	//Registers    [31]uint32
	FlagRegister uint32
}

const (
	CF  uint32 = 0b1000 // Carry Flag
	OVF uint32 = 0b0100 // Overflow Flag
	ZF  uint32 = 0b0001 // Zero Flag
	SF  uint32 = 0b0010 // Sign Flag
)

func SignExtend(val uint32) int64 {
	return int64(int32(val))
}

func SignBit(val uint32) uint32 {
	return val & 0x80000000
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

// Return true if the Zero Flag is set
func (alu *ALU) GetZF() bool {
	return (alu.FlagRegister & ZF) != 0
}

// Return true if the Sign Flag is set
func (alu *ALU) GetSF() bool {
	return (alu.FlagRegister & SF) != 0
}

// Return true if the Carry Flag is set
func (alu *ALU) GetCF() bool {
	return (alu.FlagRegister & CF) != 0
}

// Return true if the Overflow Flag is set
func (alu *ALU) GetOVF() bool {
	return (alu.FlagRegister & OVF) != 0
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

	overflow := (SignBit(sum) ^ SignBit(addend)&(SignBit(sum)^SignBit(augend))) == 1
	if overflow {
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

	overflow := (SignBit(diff)^SignBit(minuend))&(SignBit(diff)^SignBit(subtrahend)) == 1
	if overflow {
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
	quo, _ := bits.Div32(0, dividend, divisor)
	return quo
}

func (alu *ALU) Rem(dividend, divisor uint32) uint32 {
	_, rem := bits.Div32(0, dividend, divisor)
	return rem
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
	res := -a
	if res == 0 {
		alu.SetFlag(ZF)
	}
	if res&(0x1<<31) == 1 {
		// if the sign bit is set, set the sign flag
		alu.SetFlag(SF)
	}
	return res
}

func (alu *ALU) ShiftLogicalRightCarry(v uint32, amt uint32) uint32 {
	// No shift needed
	if amt == 0 {
		return v
	}
	amt = amt & 0x1F // Limit the shift amount to 0-31
	// Check if the last bit to be shifted out is 1
	lastBitShifted := (v >> (amt - 1)) & 0x1
	alu.FlagRegister |= (CF & lastBitShifted)
	return v >> amt
}

func (alu *ALU) ShiftArithRightCarry(v uint32, amt uint32) uint32 {
	// No shift needed
	if amt == 0 {
		return v
	}
	amt = amt & 0x1F // Limit the shift amount to 0-31
	// Check if the last bit to be shifted out is 1
	lastBitShifted := (int32(v) >> (amt - 1)) & 0x1
	alu.FlagRegister |= (CF & uint32(lastBitShifted))
	return uint32(int32(v) >> amt)
}

func (alu *ALU) ShiftLogicalLeftCarry(v uint32, amt uint32) uint32 {
	// No shift needed
	if amt == 0 {
		return v
	}
	amt = amt & 0x1F // Limit the shift amount to 0-31
	// Check if the last bit to be shifted out is 1
	lastBitShifted := (v >> (32 - amt)) & 0x1
	alu.FlagRegister |= (CF & lastBitShifted)
	return v << amt
}

func (alu *ALU) RotateLeft(v uint32, amt int32) uint32 {
	return bits.RotateLeft32(v, int(amt))
}
