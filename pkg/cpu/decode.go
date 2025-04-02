package cpu

import (
	"github.com/leon332157/risc-y-8/pkg/types"
	"fmt"
)
type DecodeStage struct {
	pipeline *Pipeline // Reference to the pipeline instance
	currentInstruction *InstructionIR
	next *ExecuteStage
	prev *FetchStage
}

func (d *DecodeStage) Init(pipeline *Pipeline, next Stage, prev Stage) error {
	if pipeline == nil {
		panic("[DecodeStage Init] pipeline is null")
	}
	d.pipeline = pipeline
	if next == nil {
		panic("[DecodeStage Init] next is null")
	}
	n,ok:= next.(*ExecuteStage) 
	if !ok {
		return fmt.Errorf("[fetch Init] next stage is not execute stage")
	}
	if n == nil {
		return fmt.Errorf("[fetch Init] next stage is null")
	}
	d.next = n
	p, ok := prev.(*FetchStage)
	if !ok {
		return fmt.Errorf("[fetch Init] prev stage is not fetch stage")
	}
	if p == nil {
		return fmt.Errorf("[fetch Init] prev is null")
	}
	d.prev = p
	return nil
}

func (d *DecodeStage) Name() string {
	return "Decode"
}

func (d *DecodeStage) Execute() {
	// Decode the instruction and prepare it for execution
	// This could include extracting fields from the instruction, setting control signals, etc.
	// For example:
	// d.instructionIR = DecodeInstruction(d.instruction)
	// d.next.Advance(d.instructionIR, false) // Pass the decoded instruction to the next stage
	if (d.currentInstruction == nil) {
		fmt.Println("[DecodeStage Execute] No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		return
	}
	baseInstruction := types.BaseInstruction{} // Create a new BaseInstruction to decode the instruction
	if d.currentInstruction.rawInstruction == 0 {
		panic("[DecodeStage Execute] currentRawInstruction is zero, cannot decode")
	}
	(&baseInstruction).Decode(d.currentInstruction.rawInstruction) // Decode the raw instruction into a BaseInstruction
	d.currentInstruction.BaseInstruction = baseInstruction // Store the base instruction in the InstructionIR
	
	switch baseInstruction.OpType {
	case types.RegImm:
		d.pipeline.cpu.blockRegister(baseInstruction.Rd) // block the register for writeback, if applicable
		d.currentInstruction.Op1 = d.pipeline.cpu.IntRegisters[baseInstruction.Rs].Value
		d.currentInstruction.Op2 = uint32(baseInstruction.Imm) // sign extend immediate value
		d.currentInstruction.ALUOp = baseInstruction.ALU
		d.currentInstruction.WriteBack = baseInstruction.Rd != 0 // Only write back if Rd is not zero

	case types.RegReg:
		d.pipeline.cpu.blockRegister(baseInstruction.Rd)
		d.currentInstruction.ALUOp = baseInstruction.ALU
		d.currentInstruction.Op1 = d.pipeline.cpu.IntRegisters[baseInstruction.Rs].Value
		d.currentInstruction.Op2 = d.pipeline.cpu.IntRegisters[baseInstruction.Rd].Value
		d.currentInstruction.WriteBack = baseInstruction.Rd != 0 // Only write back if Rd is not zero, otherwise it might be a nop operation

	case types.Control:
		d.currentInstruction.ControlFlag = baseInstruction.Flag
		d.currentInstruction.ControlMode = baseInstruction.Mode
		d.currentInstruction.Op1 = d.pipeline.cpu.IntRegisters[baseInstruction.RMem].Value
		d.currentInstruction.Op2 = uint32(baseInstruction.Imm) // sign extend immediate value
		d.currentInstruction.WriteBack = true

	case types.LoadStore:
		d.pipeline.cpu.blockRegister(baseInstruction.Rd)
		d.currentInstruction.MemOp = baseInstruction.Mode
		d.currentInstruction.Op1 = d.pipeline.cpu.IntRegisters[baseInstruction.RMem].Value
		d.currentInstruction.Op2 = uint32(baseInstruction.Imm) // sign extend immediate value
		d.currentInstruction.WriteBack = true

	}
}

func (d *DecodeStage) Advance(i *InstructionIR, stalled bool) {
	if stalled {
		fmt.Printf("[%v] previous stage %v returned stall\n", d.Name(), d.prev.Name())
		//d.next.Advance(nil, true)
		//return
	}
	fmt.Printf("[%v] Advancing to next stage with instruction: %+v\n", d.Name(), d.currentInstruction)
	d.next.Advance(d.currentInstruction, false) // Pass the instruction to the next stage
	d.currentInstruction = i
}	

