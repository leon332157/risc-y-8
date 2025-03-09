package testing
import (
	"testing"
	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/types"
)

func newTestCPU() *alu.CPU {
    return &alu.CPU{
        IntRegisters: [16]uint32{},
        RFlag:        0,
    }
}

func TestADD(t *testing.T) {
    cpu := newTestCPU()
    // Set r1 to 10 and add 5 (so result must be 15)
    cpu.IntRegisters[0] = 10
    cpu.ADD(1, 5) // ADD r1, 5

    if cpu.IntRegisters[0] != 15 {
        t.Errorf("ADD failed: expected r1 to be 15, got %d", cpu.IntRegisters[0])
    }
    // Expect no zero flag on a non-zero result
    if cpu.CheckFlag(types.ZF) {
        t.Errorf("ADD failed: unexpected Zero Flag set")
    }
}

func TestSUB(t *testing.T) {
    cpu := newTestCPU()
    // Set r1 to 20 and perform SUB where rd = imm - rd: r1 becomes 5 - 20 = -15 in two's complement.
    cpu.IntRegisters[0] = 20
    cpu.SUB(1, 5) // SUB r1, 5

    // Calculate two's complement result of (5 - 20)
	expected := uint32(0xFFFFFFF1)
    if cpu.IntRegisters[0] != expected {
        t.Errorf("SUB failed: expected r1 to be %d, got %d", expected, cpu.IntRegisters[0])
    }
    // We expect the sign flag (SF) to be set for a negative result.
    if !cpu.CheckFlag(types.SF) {
        t.Errorf("SUB failed: expected Sign Flag to be set")
    }
}

func TestMUL(t *testing.T) {
    cpu := newTestCPU()
    // Set r1 to 3 and multiply by 7 to get 21
    cpu.IntRegisters[0] = 3
    cpu.MUL(1, 7) // MUL r1, 7

    if cpu.IntRegisters[0] != 21 {
        t.Errorf("MUL failed: expected r1 to be 21, got %d", cpu.IntRegisters[0])
    }
    // Flags: if multiplication result is not zero, ZF should be clear.
    if cpu.CheckFlag(types.ZF) {
        t.Errorf("MUL failed: unexpected Zero Flag set")
    }
}

func TestEncodeInstructions(t *testing.T) {
	inst1 := types.Instruction{		// ADD r1, r2
		DataType: 	types.Integer,	// 01
        OpType: 	types.RegReg,	// 01
        Rd:     	1,				// 00001
        Opcode: 	types.ADD_RR,	// 0000
        Rs:     	2,				// 00010
    }

    inst2 := types.Instruction{		// LDW r3, [r4 + 0x100]
		DataType: types.Integer,	// 01
        OpType: types.LoadStore,	// 10
        Rd:     3,					// 00011
        Opcode: types.LDW,			// 00
        Rs:     4,					// 00100
        Imm:    0x100,				// 0000 0001 0000 0000
    }

    inst3 := types.Instruction{		// BEQ [r5 + 0x200]
		DataType: types.Integer,	// 01
        OpType: types.Control,		// 11
        Rs:     5,					// 00101
        Opcode: types.ControlOps[0],// BEQ (Flag 0000, Mode 000)
        Imm:    0x200,				// 0000 0000 0010 0000
    }

	inst4 := types.Instruction{		// BOF [r6 + 0x400]
		DataType: types.Integer,	// 01
		OpType: types.Control,		// 11
		Rs:     6,					// 00110
		Opcode: types.ControlOps[7],// BOF (Flag 0100, Mode 001)
		Imm:    0x400,  			// 0000 0100 0000 0000
	}

	inst5 := types.Instruction{		// LDI r11, 0x12345
		DataType: types.Integer,	// 01
		OpType: types.RegImm,		// 00
		Rd:     11,					// 01011
		Opcode: types.LDI,			// 1100
		Imm:    0x12345,  			// 0010 0100 0110 1000 101
	}

	tests := []struct {
		name 		string
		inst 		types.Instruction
		expected 	uint32
	}{
		{"inst1", inst1, 0x00004015},
		{"inst2", inst2, 0x01002039},
		{"inst3", inst3, 0x0200005D},
		{"inst4", inst4, 0x0400426D},
		{"inst5", inst5, 0x2468B8B1},
	}

	for _, tt := range tests {
		
		if encoded := tt.inst.Encode(); encoded != tt.expected {
			t.Errorf("%s: expected 0x%08X, got 0x%08X", tt.name, tt.expected, encoded)
		}
	}
}

func TestDecodeInstructions(t *testing.T) {

	tests := []struct {
		name 		string
		encoded 	uint32
		expected 	types.Instruction
	}{
		{"inst1", 0x00004015, types.Instruction{DataType: types.Integer, OpType: types.RegReg, Rd: 1, Opcode: types.ADD_RR, Rs: 2}},
		{"inst2", 0x01002039, types.Instruction{DataType: types.Integer, OpType: types.LoadStore, Rd: 3, Opcode: types.LDW, Rs: 4, Imm: 0x100}},
		{"inst3", 0x0200005D, types.Instruction{DataType: types.Integer, OpType: types.Control, Rs: 5, Opcode: types.ControlOps[0], Imm: 0x200}},
		{"inst4", 0x0400426D, types.Instruction{DataType: types.Integer, OpType: types.Control, Rs: 6, Opcode: types.ControlOps[7], Imm: 0x400}},
		{"inst5", 0x2468B8B1, types.Instruction{DataType: types.Integer, OpType: types.RegImm, Rd: 11, Opcode: types.LDI, Imm: 0x12345}},
	}

	for _, tt := range tests {
		
		if decoded := types.Decode(tt.encoded); decoded != tt.expected {
			t.Errorf("%s: expected %+v, got %+v", tt.name, tt.expected, decoded)
		}
	}
}

func main() {

	// Run tests
	t := &testing.T{}
	TestADD(t)
	TestSUB(t)
	TestMUL(t)
	TestEncodeInstructions(t)
	TestDecodeInstructions(t)

	if !t.Failed() {
		println("All tests passed!")
	}

}