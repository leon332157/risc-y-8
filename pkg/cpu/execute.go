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
	state int
	cyclesLeft uint
}

const (
	EXEC_free = iota
	EXEC_busy_int // busy waiting for integer alu to finish
	EXEC_busy_float // busy waiting for fpu to finish
	EXEC_busy_vector // busy waiting for vector unit to finish
)

func (e *ExecuteStage) Init(pipeline *Pipeline, next Stage, prev Stage) error {
	if pipeline == nil {
		e.pipeline.log.Fatal().Msg("[Execute Init] pipeline is null")
	}
	e.pipeline = pipeline
	if next == nil {
		e.pipeline.log.Fatal().Msg("[Execute Init] next is null")
	}
	n, ok := next.(*MemoryStage)
	if !ok {
		e.pipeline.log.Fatal().Msg("[Execute Init] next stage is not memory stage")
	}
	if n == nil {
		e.pipeline.log.Fatal().Msg("[Execute Init] next stage is null")
	}
	e.next = n
	p, ok := prev.(*DecodeStage)
	if !ok {
		e.pipeline.log.Fatal().Msg("[Execute Init] prev stage is not decode stage")
	}
	if p == nil {
		e.pipeline.log.Fatal().Msg("[Execute Init] prev is null")
	}
	e.prev = p
	return nil
}

func (e *ExecuteStage) Name() string {
	return "Execute"
}

func (e *ExecuteStage) ALURI() {
	// Perform ALU operation for RegImm type instruction
	inst := e.currentInstruction
	op1 := inst.Result // First operand (from a register)
	op2 := inst.Operand // Second operand (immediate value)

	// operations are done 
	switch inst.BaseInstruction.ALU {
	case types.IMM_ADD:
		inst.Result = e.pipeline.cpu.ALU.Add(op2, op2)
	case types.IMM_SUB:
		inst.Result = e.pipeline.cpu.ALU.Sub(op2, op1)
	case types.IMM_MUL:
		inst.Result = e.pipeline.cpu.ALU.Mul(op1, op2)
	case types.IMM_AND:
		inst.Result = e.pipeline.cpu.ALU.And(op1, op2)
	case types.IMM_XOR:
		inst.Result = e.pipeline.cpu.ALU.Xor(op1, op2)
	case types.IMM_OR:
		inst.Result = e.pipeline.cpu.ALU.Or(op1, op2)
	case types.IMM_NOT:
		inst.Result = e.pipeline.cpu.ALU.Not(op1)
	case types.IMM_NEG:
		inst.Result = e.pipeline.cpu.ALU.Neg(op1)
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
	switch instructionIR.BaseInstruction.MemMode {

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
	instruction.WriteBack = true
	if instruction.ControlFlag == types.EQ.Flag && instruction.ControlMode == types.EQ.Mode {
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
		fmt.Println("[ExecuteStage Execute] No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		return
	}
	if e.state != EXEC_free {
		e.cyclesLeft--
		e.pipeline.sTraceF(e, "busy %v cycles left %v", e.state, e.cyclesLeft)
		if e.cyclesLeft == 0 {
			e.state = EXEC_free
		}
	}
	switch e.currentInstruction.BaseInstruction.OpType {
	case types.RegImm:
		e.ALURI()
	case types.RegReg:
		e.ALURR() // Perform ALU operation for RegReg type instruction
	case types.LoadStore:
		e.LoadStore() // Handle Load/Store operations, passed to memory stage for the actual operation
	case types.Control:
		e.Control() // Handle Control operations, this function should be implemented in the memory stage or similar
	default:
		fmt.Println(e.currentInstruction.BaseInstruction.OpType)
		panic("unsupported instruction type in Execute stage") // Handle unsupported instruction types
	}
}
func (e *ExecuteStage) Advance(i *InstructionIR, stalled bool) {
	if stalled {
		fmt.Printf("[%v] previous stage %v returned stall\n", e.Name(), e.prev.Name())
		//e.next.Advance(nil, true)
		//return
	}
	fmt.Printf("[%v] Advancing to next stage with instruction: %+v\n", e.Name(), e.currentInstruction)
	e.next.Advance(e.currentInstruction, false) // Pass the instruction to the next stage
	e.currentInstruction = i
}

func (e *ExecuteStage) CanAdvance() bool {
	return e.state == EXEC_free
}
