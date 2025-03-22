package grammar

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestBasic(t *testing.T) {
	example :=
		`
		.org 0x1
		#xor r1, r1
		hlt
		.L1:
		nop
		add r1, 1
		xor r2, r2
		add r2, -2
		add r1, r2
		ldr r2, [r1+10]
		ldr r2, [r1-0x20]
		bunc [r13]
		cmp r2, 0x10
		nop
		xor r3,r3
		orr r3, 0x3f80
		shl r3, 16
		push r3
		fldr f1, [sp]
		pop r3
		xor r3, r3
		add r3, 0x4
		shl r3, 28
		push r3
		fldr f2, [sp]
		fadd f3, f1, f2
		fmul f4, f3, f2
		fsqrt f5, f4
		bunc 0x1000
		hlt
		VZEROALL v0
		VZEROUPPER v1
		VBEQ v2, v3, [pc-10]
		VLD v8,[r1+0xFFFF]
	`

	prog, err := ParseString("example.asm", example)
	if err != nil {
		t.Errorf("error %v\n", err)
	}
	for _, each := range prog.Lines {
		t.Logf("line: %+v\n", each)
		if each.Label != nil {
			t.Logf("label: %+v\n", each.Label)
		}
		if each.Directive != nil {
			t.Logf("directive: %+v\n", each.Directive)
		}
		if each.Instruction != nil {
			t.Logf("instr: %+v\n", each.Instruction)
		}
	}
}


func TestRI(t *testing.T) {
	var example = "add r1,1"
	prog, err := ParseString("testRI1", example)
	if err != nil {
		t.Errorf("[TestRI] error %v\n", err)
	}
	t.Logf("[TestRI] prog.Lines: %+v\n", prog)
	var expected = Instruction{
		Mnemonic: "add",
		Operands: []Operand{OperandRegister{Value:"r1"}, OperandImmediate{Value:"1"}},
	}
	t.Logf("[TestRI] log: %+v\n",cmp.Equal(*(prog.Lines[0].Instruction), expected))
	//if !cmp.Equal(*(prog.Lines[0].Instruction), expected) {
	//	t.Errorf("[TestRI] prog.Lines[0] = %+v\n !=\n %+v\n",prog.Lines[0].Instruction, expected)
	//}
}

func testRR(t *testing.T) {
	var example = "add r1,r2"
	prog, err := ParseString("testRR", example)
	if err != nil {
		t.Errorf("[TestRR] error %v\n", err)
	}
	t.Logf("[TestRR] prog.Lines: %+v\n", prog)
	var expected = Instruction{
		Mnemonic: "add",
		Operands: []Operand{OperandRegister{Value:"r1"}, OperandRegister{Value:"r2"}},
	}
	if !cmp.Equal(*(prog.Lines[0].Instruction), expected) {
		t.Errorf("[TestRR] prog.Lines[0] = %+v\n !=\n %+v\n", prog.Lines[0].Instruction, expected)
	}
}