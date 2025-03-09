package alu
import (
	"fmt"
	"github.com/leon332157/risc-y-8/pkg/types"
)

func main() {

	inst1 := types.Instruction{			// ADD r1, r2
		DataType: 	types.Integer,		// 01
		OpType: 	types.RegReg,		// 01
		Rd:     	1,					// 00001
		Opcode: 	types.ADD_RR,		// 0000
		Rs:     	2,					// 00010
	}

	inst2 := types.Instruction{			// LDW r3, [r4 + 0x100]
		DataType: types.Integer,		// 01
		OpType: types.LoadStore,		// 10
		Rd:     3,						// 00011
		Opcode: types.LDW,				// 00
		Rs:     4,						// 00100
		Imm:    0x100,					// 0000 0001 0000 0000
	}

	inst3 := types.Instruction{			// BEQ [r5 + 0x200]
		DataType: types.Integer,		// 01
		OpType: types.Control,			// 11
		Rs:     5,						// 00101
		Opcode: types.ControlOps[0],	// BEQ (Flag 0000, Mode 000)
		Imm:    0x200,					// 0000 0000 0010 0000
	}

	inst4 := types.Instruction{			// BOF [r6 + 0x400]
		DataType: types.Integer,		// 01
		OpType: types.Control,			// 11
		Rs:     6,						// 00110
		Opcode: types.ControlOps[7], 	// BOF (Flag 0100, Mode 001)
		Imm:    0x400,  				// 0000 0100 0000 0000
	}

	inst5 := types.Instruction{			// LDI r11, 0x12345
		DataType: types.Integer,		// 01
		OpType: types.RegImm,			// 00
		Rd:     11,						// 01011
		Opcode: types.LDI,				// 1100
		Imm:    0x12345,  				// 0010 0100 0110 1000 101
	}

    fmt.Printf("Instruction 1: 0x%08X should be 0x00004015\n", inst1.Encode())
	fmt.Printf("Instruction 2: 0x%08X should be 0x01002039\n", inst2.Encode())
	fmt.Printf("Instruction 3: 0x%08X should be 0x0200005D\n", inst3.Encode())
	fmt.Printf("Instruction 4: 0x%08X should be 0x0400426D\n", inst4.Encode())
	fmt.Printf("Instruction 5: 0x%08X should be 0x2468B8B1\n", inst5.Encode())
}