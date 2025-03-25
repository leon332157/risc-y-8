package assembler

func (inst *BaseInstruction) Encode() uint32 {

	var encoded uint32 = 0

	// Common fields for all instructions
	encoded |= 0b01                     // Bits 1-0 (DataType)
	encoded |= uint32(inst.OpType) << 2 // Bits 3-2 (OpType)

	switch inst.OpType {

	case RegImm:

		encoded |= uint32(inst.Rd) << 4   // 5 bit Rd (Bits 9-4)
		encoded |= uint32(inst.ALU) << 9  // 4 bit ALU Op (Bits 12-9)
		encoded |= uint32(inst.Imm) << 16 // 16 bit Immediate (Bits 31-13)

	case RegReg:

		encoded |= uint32(inst.Rd) << 4  // 5 bit Rd (Bits 9-4)
		encoded |= uint32(inst.ALU) << 9 // 4 bit ALU Op (Bits 12-9)
		encoded |= uint32(inst.Rs) << 13 // 5 bit Rs (Bits 18-14)

	case LoadStore:

		encoded |= uint32(inst.Rd) << 4    // 5 bit Rd (Bits 9-4)
		encoded |= uint32(inst.Mode) << 9  // 2 bit Mode (Bits 11-10)
		encoded |= uint32(inst.RMem) << 11 // 5 bit Rs (Bits 16-12)
		encoded |= uint32(inst.Imm) << 16  // 16 bit Immediate (Bits 31-20)

	case Control:

		encoded |= uint32(inst.RMem) << 4  // 5 bit Rd (Bits 9-4)
		encoded |= uint32(inst.Flag) << 9  // 4 bit flag (Bits 12-9)
		encoded |= uint32(inst.Mode) << 13 // 3 bit Mode (Bits 16-13)
		encoded |= uint32(inst.Imm) << 16  // 16 bit Immediate (Bits 32-16)

	}

	return encoded
}
