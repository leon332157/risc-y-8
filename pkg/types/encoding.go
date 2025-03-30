package types

func (inst *BaseInstruction) Encode() uint32 {

	var encoded uint32 = 0

	// Common fields for all instructions
	encoded |= 0b01                     // Bits 1-0 (DataType)
	encoded |= uint32(inst.OpType) << 2 // Bits 3-2 (OpType)

	switch inst.OpType {

	case RegImm:

		encoded |= uint32(inst.Rd) << 4   // 5 bit Rd (Bits 8-4)
		encoded |= uint32(inst.ALU) << 9  // 4 bit ALU Op (Bits 12-9)
		encoded |= uint32(0) << 13		  // 3 bit reserve (Bits 15-13)
		encoded |= uint32(inst.Imm) << 16 // 16 bit Immediate (Bits 31-16)

	case RegReg:

		encoded |= uint32(inst.Rd) << 4  // 5 bit Rd (Bits 8-4)
		encoded |= uint32(inst.ALU) << 9 // 4 bit Opcode (Bits 12-9)
		encoded |= uint32(inst.Rs) << 13 // 5 bit Rs (Bits 17-13)

	case LoadStore:

		encoded |= uint32(inst.Rd) << 4    // 5 bit Rd (Bits 8-4)
		encoded |= uint32(inst.Mode) << 9  // 2 bit Mode (Bits 10-9)
		encoded |= uint32(inst.RMem) << 11 // 5 bit RMem (Bits 15-11)
		encoded |= uint32(inst.Imm) << 16  // 16 bit Immediate (Bits 31-16)

	case Control:

		encoded |= uint32(inst.RMem) << 4  // 5 bit RMem (Bits 8-4)
		encoded |= uint32(inst.Flag) << 9  // 4 bit Flag (Bits 12-9)
		encoded |= uint32(inst.Mode) << 13 // 3 bit Mode (Bits 15-13)
		encoded |= uint32(inst.Imm) << 16  // 16 bit Immediate (Bits 31-16)

	}

	return encoded
}

func (inst *BaseInstruction) Decode(encoded uint32) {
	// Bits 1-0 (DataType) , ignored since this is a base instruction, always has DataType 0b01
	inst.OpType = uint8((encoded >> 2) & 0b11) // Bits 3-2 (OpType)
	switch inst.OpType {

	case RegImm:

		inst.Rd = uint8((encoded >> 4) & 0x1F)     // 5 bit Rd (Bits 8-4)
		inst.ALU = uint8((encoded >> 9) & 0xF)     // 4 bit ALU Op (Bits 12-9)
		inst.Imm = int16((encoded >> 16) & 0xFFFF) // 16 bit Immediate (Bits 28-13)

	case RegReg:

		inst.Rd = uint8((encoded >> 4) & 0x1F)  // 5 bit Rd (Bits 8-4)
		inst.ALU = uint8((encoded >> 9) & 0xF)  // 4 bit Opcode (Bits 12-9)
		inst.Rs = uint8((encoded >> 13) & 0x1F) // 5 bit Rs (Bits 17-13)

	case LoadStore:

		inst.Rd = uint8((encoded >> 4) & 0x1F)     // 5 bit Rd (Bits 8-4)
		inst.Mode = uint8((encoded >> 9) & 0x3)    // 2 bit Mode (Bits 10-9)
		inst.RMem = uint8((encoded >> 11) & 0x1F)  // 5 bit RMem (Bits 15-11)
		inst.Imm = int16((encoded >> 16) & 0xFFFF) // 16 bit Immediate (Bits 31-16)

	case Control:

		inst.RMem = uint8((encoded >> 4) & 0x1F)   // 5 bit RMem (Bits 8-4)
		inst.Flag = uint8((encoded >> 9) & 0xF)    // 4 bit Flag (Bits 12-9)
		inst.Mode = uint8((encoded >> 13) & 0x7)   // 3 bit Mode (Bits 15-13)
		inst.Imm = int16((encoded >> 16) & 0xFFFF) // 16 bit Immediate (Bits 31-16)

	default:
		panic("Invalid OpType")
	}
}
