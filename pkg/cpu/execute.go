package cpu

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/types"
)

type ExecuteStage struct {
	pipeline *Pipeline    // Reference to the pipeline instance
	next     *MemoryStage // Next stage in the pipeline
	prev     *DecodeStage // Previous stage in the pipeline
	//currentRawInstruction  uint32       // Instruction to be executed
	currentInstruction *InstructionIR
}

func (e *ExecuteStage) Init(pipeline *Pipeline, next Stage, prev Stage) error {
	if pipeline == nil {
		panic("[DecodeStage Init] pipeline is null")
	}
	e.pipeline = pipeline
	if next == nil {
		panic("[DecodeStage Init] next is null")
	}
	n, ok := next.(*MemoryStage)
	if !ok {
		return fmt.Errorf("[fetch Init] next stage is not memory stage")
	}
	if n == nil {
		return fmt.Errorf("[fetch Init] next stage is null")
	}
	e.next = n
	p, ok := prev.(*DecodeStage)
	if !ok {
		return fmt.Errorf("[fetch Init] prev stage is not decode stage")
	}
	if p == nil {
		return fmt.Errorf("[fetch Init] prev is null")
	}
	e.prev = p
	return nil
}

func (e *ExecuteStage) Name() string {
	return "Execute"
}

func (e *ExecuteStage) ALU(instructionIR *InstructionIR) {
	// Perform ALU operation for RegImm type instruction
	op1 := instructionIR.Op1 // First operand (from a register)
	op2 := instructionIR.Op2 // Second operand (immediate value)
	switch instructionIR.ALUOp {
	case types.IMM_ADD:
		instructionIR.Result = e.pipeline.cpu.ALU.Add(op1, op2)
	case types.IMM_SUB: // since add and sub overlap, we can use the same case
		instructionIR.Result = e.pipeline.cpu.ALU.Add(op1, op2)
	case types.IMM_MUL:
		instructionIR.Result = e.pipeline.cpu.ALU.Mul(op1, op2)
	case types.IMM_CMP:
	case types.REG_CMP:
		instructionIR.WriteBack = false
		e.pipeline.cpu.ALU.Sub(op1, op2)
	default:
		panic("unsupported ALU operation for RegImm type instruction") // Handle unsupported ALU operations
		// Handle other ALU operations as needed
	}
}

func (e *ExecuteStage) calculateMemAddr(inst *InstructionIR) uint32 {
	if inst == nil {
		panic("instructionIR is nil")
	}
	return (inst.Op1 + inst.Op2) % uint32(e.pipeline.cpu.RAM.SizeLines()) // Calculate the memory address for load/store instructions based on the operands
}

func (e *ExecuteStage) LoadStore(instructionIR *InstructionIR) {
	switch instructionIR.BaseInstruction.Mode {

	case LOAD:
		// Load instruction
		instructionIR.DestMemAddr = e.calculateMemAddr(instructionIR)
	case STORE:
		// Store instruction
		instructionIR.DestMemAddr = e.calculateMemAddr(instructionIR)
		instructionIR.Result = instructionIR.Op1
	default:
		panic("Push pop not implemented")
	}
}

func (e *ExecuteStage) Control(instruction *InstructionIR) {
	// Handle branch instructions, if applicable
	if instruction.BaseInstruction.OpType != types.Control {
		panic("not a control instruction")
	}
	instruction.WriteBack = true
	if (instruction.ControlFlag == types.EQ.Flag && instruction.ControlMode == types.EQ.Mode) {
		if (e.pipeline.cpu.Flag & 0b1) == 0 {
			// take the branch
			instruction.DestMemAddr = e.calculateMemAddr(instruction)
			instruction.BaseInstruction.Rd = 0
		} else {
			instruction.BaseInstruction.Rd = 0x1F // FIXME: Don't do this
			// do not take the branch
			// fmt.Println("not taking branch")
			return // Do not take the branch, exit early
		}
	}

}

func (e *ExecuteStage) Execute() {
	if e.currentInstruction == nil {
		// If there is no instructionIR to execute, return early
		return
	}
	switch e.currentInstruction.BaseInstruction.OpType {
	case types.RegImm:
		e.ALU(e.currentInstruction)
	case types.RegReg:
		e.ALU(e.currentInstruction) // Perform ALU operation for RegReg type instruction
	case types.LoadStore:
		e.LoadStore(e.currentInstruction) // Handle Load/Store operations, this function should be implemented in the memory stage or similar
	case types.Control:
		e.Control(e.currentInstruction) // Handle Control operations, this function should be implemented in the memory stage or similar
	default:
		fmt.Println(e.currentInstruction.BaseInstruction.OpType)
		panic("unsupported instruction type in Execute stage") // Handle unsupported instruction types
	}
}
func (e *ExecuteStage) Advance(i *InstructionIR, stalled bool) {
	if stalled {
		fmt.Printf("[%v] previous stage %v returned stall ", e.Name(), e.prev.Name())
		e.next.Advance(nil, true)
		return
	}
	e.next.Advance(e.currentInstruction, false) // Pass the instruction to the next stage
	e.currentInstruction = i
}
