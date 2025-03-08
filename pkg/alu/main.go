package main
import "fmt"

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
	integer		DataType = 0b01
	float		DataType = 0b11
	vector		DataType = 0b11
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
	Mode uint8	// 2 bits
	Flag uint8	// 5 bits
	Name string
}

var ControlOps = []ControlOp{
    {0b000, 0b0000, "BEQ"},
    {0b000, 0b0001, "BNE"},
    {0b000, 0b0101, "BLT"},
    {0b011, 0b0110, "BGE"},
    {0b100, 0b1000, "BLU"},
    {0b000, 0b1000, "BAE"},
    {0b010, 0b1011, "BA"},
    {0b001, 0b0100, "BOF"},
    {0b000, 0b0100, "BNF"},
	{0b111, 0b1111, "BUNC"},
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

func main() {
    inst1 := Instruction{		// ADD r1, r2
		DataType: 	integer,	// 01
        OpType: 	RegReg,		// 01
        Rd:     	1,			// 00001
        Opcode: 	ADD_RR,		// 0000
        Rs:     	2,			// 00010
    }

    inst2 := Instruction{		// LDW r3, [r4 + 0x100]
		DataType: integer,		// 01
        OpType: LoadStore,		// 10
        Rd:     3,				// 00011
        Opcode: LDW,			// 00
        Rs:     4,				// 00100
        Imm:    0x100,			// 0000 0001 0000 0000
    }

    inst3 := Instruction{		// BEQ [r5 + 0x200]
		DataType: integer,		// 01
        OpType: Control,		// 11
        Rs:     5,				// 00101
        Opcode: ControlOps[0],	// BEQ (Flag 0000, Mode 000)
        Imm:    0x200,			// 0000 0000 0010 0000
    }

	inst4 := Instruction{		// BOF [r6 + 0x400]
		DataType: integer,		// 01
		OpType: Control,		// 11
		Rs:     6,				// 00110
		Opcode: ControlOps[7], 	// BOF (Flag 0100, Mode 001)
		Imm:    0x400,  		// 0000 0100 0000 0000
	}

	inst5 := Instruction{		// LDI r11, 0x12345
		DataType: integer,		// 01
		OpType: RegImm,			// 00
		Rd:     11,				// 01011
		Opcode: LDI,			// 1100
		Imm:    0x12345,  		// 0010 0100 0110 1000 101
	}

    fmt.Printf("Instruction 1: 0x%08X should be 0x00004015\n", inst1.Encode())
	fmt.Printf("Instruction 2: 0x%08X should be 0x01002039\n", inst2.Encode())
	fmt.Printf("Instruction 3: 0x%08X should be 0x0200005D\n", inst3.Encode())
	fmt.Printf("Instruction 4: 0x%08X should be 0x0400426D\n", inst4.Encode())
	fmt.Printf("Instruction 5: 0x%08X should be 0x2468B8B1\n", inst5.Encode())
}