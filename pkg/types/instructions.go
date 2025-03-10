package types
import (
	"fmt"
)

type OpType uint8
const (
	RegImm		OpType = 0b00
	RegReg		OpType = 0b01
	LoadStore	OpType = 0b10
	Control		OpType = 0b11
)

type DataType uint8
const (
	// none		DataType = 0b00
	Integer		DataType = 0b01
	Float		DataType = 0b11
	Vector		DataType = 0b11
)

// Register to Immediate | OpType 00 | DataType 01
type ALUOp uint8

const (
	ADD	ALUOp = iota
	SUB
	MUL
	AND
	XOR
	ORR
	NOT
	NEG
	SHR
	SAR
	SHL
	ROL
	LDI
	LDX
	CMP
)

// Register to Register | OpType 01 | DataType 01
type RegRegOp uint8

const (
	ADD_RR	RegRegOp = iota
	SUB_RR
	MUL_RR
	DIV_RR
	ORR_RR
	XOR_RR
	NOT_RR
	AND_RR
	SHL_RR
	SHR_RR
	SAR_RR
	ROL_RR
	CMP_RR
	CPY_RR
)

// Load Store | OpType 10 | DataType 01
type LoadStoreOp uint8
const (
	LDW	LoadStoreOp = iota
	POP
	PUSH
	STW
)

// Control | OpType 11 | DataType 01
type ControlOp struct {
	Mode uint8	// 3 bits
	Flag uint8	// 4 bits
	Name string	// Name of the instruction
}

var ControlOps = map[string]ControlOp{
    "BEQ":  {Mode: 0b000, Flag: 0b0000, Name: "BEQ"},
    "BNE":  {Mode: 0b000, Flag: 0b0001, Name: "BNE"},
    "BLT":  {Mode: 0b000, Flag: 0b0101, Name: "BLT"},
    "BGE":  {Mode: 0b011, Flag: 0b0110, Name: "BGE"},
    "BLU":  {Mode: 0b100, Flag: 0b1000, Name: "BLU"},
    "BAE":  {Mode: 0b000, Flag: 0b1000, Name: "BAE"},
    "BA":   {Mode: 0b010, Flag: 0b1011, Name: "BA"},
    "BOF":  {Mode: 0b001, Flag: 0b0100, Name: "BOF"},
    "BNF":  {Mode: 0b000, Flag: 0b0100, Name: "BNF"},
    "BUNC": {Mode: 0b111, Flag: 0b1111, Name: "BUNC"},
}

type Instruction struct {
	DataType	DataType		// Data Type (2 bit)
	OpType		OpType			// Operation Type (2 bit)
	Opcode		interface{}		// Can be RegImm, RegReg, LoadStore, Control
	Rd			uint8			// Destination Register (5 bit)
	Rs			uint8			// Source Register (5 bit) (if applicable)
	Imm 		uint32			// Immediate or Displacement Value (if applicable)
}

func (inst Instruction) Encode() uint32 {

	var encoded uint32

	// Common fields for all instructions
	encoded |= uint32(inst.DataType) & 0b11	// Bits 1-0 (DataType)
	encoded |= (uint32(inst.OpType) & 0b11) << 2	// Bits 3-2 (OpType)

	switch inst.OpType {

	case RegImm:

		opcode, ok := inst.Opcode.(ALUOp)
		
		if !ok {
			fmt.Println("Invalid ALUOp")
			return 0
		}

		encoded |= (uint32(inst.Imm) & 0x7FFFF) << 13	// 19 bit Immediate (Bits 31-13)
		encoded |= (uint32(opcode) & 0xF) << 9		// 4 bit ALU Op (Bits 12-9)
		encoded |= (uint32(inst.Rd) & 0x1F) << 4		// 5 bit Rd (Bits 8-4)

	case RegReg:

		opcode, ok := inst.Opcode.(RegRegOp)

		if !ok {
			fmt.Println("Invalid RegRegOp")
			return 0
		}

		encoded |= (uint32(inst.Rs) & 0x1F) << 13		// 5 bit Rs (Bits 18-14)
		encoded |= (uint32(opcode) & 0xF) << 8			// 4 bit Opcode (Bits 13-9)
		encoded |= (uint32(inst.Rd) & 0x1F) << 4		// 5-bit Rd (Bits 8-4)

	case LoadStore:

		opcode, ok := inst.Opcode.(LoadStoreOp)

		if !ok {
			fmt.Println("Invalid LoadStoreOp")
			return 0
		}

		encoded |= (uint32(inst.Imm) & 0xFFFF) << 16	// 16 bit Displacement (Bits 31-16)
		encoded |= (uint32(inst.Rs) & 0x1F) << 11		// 5 bit rmem (Bits 15-11)
		encoded |= (uint32(opcode) & 0x3) << 9			// 2 bit Opcode (Mode) (Bits 10-9)
		encoded |= (uint32(inst.Rd) & 0x1F) << 4		// 5 bit Rd (Bits 8-4)

	case Control:

		opcode, ok := inst.Opcode.(ControlOp)
		
		if !ok {
			fmt.Println("Invalid ControlOp")
			return 0
		}

		encoded |= (uint32(inst.Imm) & 0xFFFF) << 16		// 16 bit Displacement (Bits 31-16)
		encoded |= (uint32(opcode.Flag) & 0xF) << 12		// 4 bit Opcode (Bits 15-12)
		encoded |= (uint32(opcode.Mode) & 0x7) << 9			// 3 bit Mode (Bits 11-9)
		encoded |= (uint32(inst.Rs) & 0x1F) << 4			// 5 bit Rs (Bits 8-4)

	}

	return encoded
}

func Decode(encoded uint32) Instruction {

	var inst Instruction

	// Extract common fields
	inst.DataType = DataType(encoded & 0b11)				// Bits 1-0 (DataType)
	inst.OpType = OpType((encoded >> 2) & 0b11)				// Bits 3-2 (OpType)

	switch inst.OpType {

	case RegImm:

		inst.Rd = uint8((encoded >> 4) & 0x1F)				// 5 bit Rd (Bits 8-4)
		inst.Opcode = ALUOp((encoded >> 9) & 0xF)			// 4 bit ALU Op (Bits 12-9)
		inst.Imm = (encoded >> 13) & 0x7FFFF				// 19 bit Immediate (Bits 31-13)

	case RegReg:

		inst.Rs = uint8((encoded >> 4) & 0x1F)				// 5 bit Rs (Bits 18-14)
		inst.Opcode = RegRegOp((encoded >> 8) & 0xF)		// 4 bit Opcode (Bits 13-9)
		inst.Rd = uint8((encoded >> 13) & 0x1F)				// 5-bit Rd (Bits 8-4)

	case LoadStore:

		inst.Rd = uint8((encoded >> 4) & 0x1F)				// 5 bit Rd (Bits 8-4)
		inst.Opcode = LoadStoreOp((encoded >> 9) & 0x3)		// Opcode (Mode) (Bits 10-9)
		inst.Rs = uint8((encoded >> 11) & 0x1F)				// rmem (Bits 15-11)
		inst.Imm = (encoded >> 16) & 0xFFFF					// Displacement (Bits 31-16)

	case Control:

		opcode := ControlOp{}

		inst.Rs = uint8((encoded >> 4) & 0x1F)				// Rs (Bits 8-4)
		opcode.Mode = uint8((encoded >>9 )&0x7)				// Mode (Bits11 - Bits9)
		opcode.Flag = uint8((encoded >>12 )&0xF)			// Opcode (Bits15 - Bits12)
		inst.Imm = (encoded >> 16) & 0xFFFF					// Displacement (Bits 31-16)

		// Reverse lookup: iterate over ControlOps map to find matching ControlOp.
		found := false
		for _, op := range ControlOps {
			if op.Flag == opcode.Flag && op.Mode == opcode.Mode {
				opcode.Name = op.Name
				found = true
				break
			}
		}
		if !found {
			fmt.Println("Warning: Unrecognized ControlOp with Flag:", opcode.Flag, "Mode:", opcode.Mode)
		}

		inst.Opcode = opcode
	}

	return inst
}