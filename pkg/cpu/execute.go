package cpu

import (
	"fmt"
	"math/bits"

	"github.com/leon332157/risc-y-8/pkg/types"
)

type ExecuteStage struct {
	pipeline *Pipeline    // Reference to the pipeline instance
	next     *MemoryStage // Next stage in the pipeline
	prev     *DecodeStage // Previous stage in the pipeline
	//currentRawInstruction  uint32       // Instruction to be executed
	currentInstruction *InstructionIR
	state              int
	cyclesLeft         uint
}

const (
	EXEC_free        = iota
	EXEC_busy_int    // busy waiting for integer alu to finish
	EXEC_busy_float  // busy waiting for fpu to finish
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
	op1 := inst.Result  // First operand (from a register)
	op2 := inst.Operand // Second operand (immediate value)

	// operations are done
	inst.WriteBack = true
	switch inst.BaseInstruction.ALU {
	case types.IMM_ADD:
		inst.Result = e.pipeline.cpu.ALU.Add(op1, op2)
	case types.IMM_SUB:
		inst.Result = e.pipeline.cpu.ALU.Sub(op1, op2)
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
	case types.IMM_SHR:
		inst.Result = e.pipeline.cpu.ALU.ShiftLogicalRightCarry(op1, op2)
	case types.IMM_SAR:
		inst.Result = e.pipeline.cpu.ALU.ShiftArithRightCarry(op1, op2)
	case types.IMM_SHL:
		inst.Result = e.pipeline.cpu.ALU.ShiftLogicalLeftCarry(op1, op2)
	case types.IMM_ROL:
		inst.Result = e.pipeline.cpu.ALU.RotateLeft(op1, int32(op2))
	case types.IMM_LDI:
		inst.Result = uint32(op2)
	case types.IMM_LDX:
		inst.Result = uint32(int32(op2))
	case types.IMM_CMP:
		e.pipeline.cpu.unblockIntR(inst.BaseInstruction.Rd)
		inst.WriteBack = false
		e.pipeline.cpu.ALU.Sub(op1, op2)
	default:
		e.pipeline.log.Panic().Msg("unsupported ALU operation for RegImm type instruction") // Handle unsupported ALU operations
		// Handle other ALU operations as needed
	}
	e.pipeline.sTracef(e, "ALURI operation result: %v", inst.Result) // For debugging purposes, log the result of the ALU operation
}

func (e *ExecuteStage) ALURR() {
	inst := e.currentInstruction
	op1 := inst.Result
	op2 := inst.Operand
	inst.WriteBack = true
	if (inst.BaseInstruction.ALU == types.REG_DIV) || (inst.BaseInstruction.ALU == types.REG_REM) {
		if op2 == 0 {
			e.pipeline.log.Panic().Msg("division by zero") // Handle division by zero error
		}
		if e.state != EXEC_free {
			e.cyclesLeft--
			return
		} else {
			e.state = EXEC_busy_int
			e.cyclesLeft = 3                                                       // Set the number of cycles for division operation
			e.pipeline.sTracef(e, "busy %v cycles left %v", e.state, e.cyclesLeft) // For debugging purposes, log the state and cycles left
			return
		}
	}
	switch inst.BaseInstruction.ALU {
	case types.REG_ADD:
		inst.Result = e.pipeline.cpu.ALU.Add(op1, op2)
	case types.REG_SUB:
		inst.Result = e.pipeline.cpu.ALU.Sub(op1, op2)
	case types.REG_MUL:
		inst.Result = e.pipeline.cpu.ALU.Mul(op1, op2)
	case types.REG_DIV:
		inst.Result = e.pipeline.cpu.ALU.Div(op1, op2)
	case types.REG_REM:
		inst.Result = e.pipeline.cpu.ALU.Rem(op1, op2)
	case types.REG_OR:
		inst.Result = e.pipeline.cpu.ALU.Or(op1, op2)
	case types.REG_XOR:
		inst.Result = e.pipeline.cpu.ALU.Xor(op1, op2)
	case types.REG_AND:
		inst.Result = e.pipeline.cpu.ALU.And(op1, op2)
	case types.REG_NOT:
		inst.Result = e.pipeline.cpu.ALU.Not(op1)
	case types.REG_SHL:
		inst.Result = e.pipeline.cpu.ALU.ShiftLogicalLeftCarry(op1, op2)
	case types.REG_SHR:
		inst.Result = e.pipeline.cpu.ALU.ShiftLogicalRightCarry(op1, op2)
	case types.REG_SAR:
		inst.Result = e.pipeline.cpu.ALU.ShiftArithRightCarry(op1, op2)
	case types.REG_ROL:
		inst.Result = e.pipeline.cpu.ALU.RotateLeft(op1, int32(op2))
	case types.REG_CMP:
		e.pipeline.cpu.unblockIntR(inst.BaseInstruction.Rd)
		inst.WriteBack = false
		e.pipeline.cpu.ALU.Sub(op1, op2)
	case types.REG_CPY:
		inst.Result = op2
	case types.REG_NSA:
		inst.Result = uint32(bits.OnesCount32(op2))
	default:
		e.pipeline.log.Panic().Msg("unsupported ALU operation for RegReg type instruction") // Handle unsupported ALU operations
	}
	e.pipeline.sTracef(e, "ALURR operation result: %v", inst.Result) // For debugging purposes, log the result of the ALU operation
}

func (e *ExecuteStage) calculateMemAddr(base uint32, displacement int32) uint32 {
	res := (base + uint32(displacement)) % uint32(e.pipeline.cpu.RAM.SizeLines()) // Calculate the memory address for load/store instructions based on the operands
	e.pipeline.sTracef(e, "calculated memory address: %v", res)                   // For debugging purposes, log the calculated memory address
	return res
}

func (e *ExecuteStage) LoadStore() {
	inst := e.currentInstruction

	switch inst.BaseInstruction.MemMode {
	case types.LDW:
		e.pipeline.sTracef(e, "ldw, calculating memory addr from %v + %v", inst.Result, int32(inst.Operand))
		inst.DestMemAddr = e.calculateMemAddr(inst.Result, int32(inst.Operand))
		inst.WriteBack = true
	case types.POP, types.PUSH:
		e.pipeline.sTrace(e, "pop/push")
		inst.RDestAux = types.IntegerRegisters["sp"]
	case types.STW:
		e.pipeline.sTracef(e, "stw, calculating memory addr from %v + %v", inst.Result, int32(inst.Operand))
		inst.DestMemAddr = e.calculateMemAddr(inst.Result, int32(inst.Operand))
		inst.WriteBack = false
	default:
		e.pipeline.log.Panic().Msg("unsupported memory operation for LoadStore type instruction")
	}
}

func combineFlags(ctrlMode, ctrlFlag uint8) uint8 {
	return ctrlMode<<4 | ctrlFlag
}

func (e *ExecuteStage) Control() {
	inst := e.currentInstruction
	combiFlag := combineFlags(inst.BaseInstruction.CtrlMode, inst.BaseInstruction.CtrlFlag)
	inst.DestMemAddr = e.calculateMemAddr(inst.DestMemAddr, int32(inst.Operand))
	alu := e.pipeline.cpu.ALU
	switch combiFlag {
	case types.GetModeFlag(types.UNC): // unconditional branch
		inst.BranchTaken = true
	case types.GetModeFlag(types.CALL): // call
		inst.BranchTaken = true
		inst.RDestAux = types.IntegerRegisters["lr"]
	case types.GetModeFlag(types.NE):
		if false == alu.GetZF() {
			// if zero flag is zero, branch
			inst.BranchTaken = true
		}
	case types.GetModeFlag(types.EQ):
		if alu.GetZF() {
			// if zero flag is set, branch
			inst.BranchTaken = true
		}
	case types.GetModeFlag(types.LT):
		if alu.GetSF() != alu.GetOVF() {
			// if sign flag is not equal to overflow flag, branch
			inst.BranchTaken = true
		}
	case types.GetModeFlag(types.GE):
		if alu.GetSF() == alu.GetOVF() { // if sign flag is equal to overflow flag, branch
			inst.BranchTaken = true
		}
	case types.GetModeFlag(types.LU):
		if alu.GetCF() {
			// if carry flag is set, branch
			inst.BranchTaken = true
		}
	case types.GetModeFlag(types.AE):
		if false == alu.GetCF() {
			// if carry flag is zero, branch
			inst.BranchTaken = true
		}
	case types.GetModeFlag(types.A):
		if false == alu.GetZF() && false == alu.GetCF() {
			// if zero flag and carry flag are both zero, branch
			inst.BranchTaken = true
		}
	case types.GetModeFlag(types.OF):
		if alu.GetOVF() {
			// if overflow flag is set, branch
			inst.BranchTaken = true
		}
	case types.GetModeFlag(types.NF):
		if false == alu.GetOVF() {
			// if overflow flag is zero, branch
			inst.BranchTaken = true
		}
	}

}

func (e *ExecuteStage) Execute() {
	if e.currentInstruction == nil {
		// If there is no instructionIR to execute, return early
		e.pipeline.sTrace(e, "No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		return
	}
	if e.state != EXEC_free {
		e.cyclesLeft--
		e.pipeline.sTracef(e, "busy %v cycles left %v", e.state, e.cyclesLeft)
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
func (e *ExecuteStage) Advance(i *InstructionIR, prevstalled bool) bool {
	if prevstalled {
		e.pipeline.sTracef(e, "previous stage %v returned stall\n", e.prev.Name())
	}
	if e.state != EXEC_free {
		e.pipeline.sTrace(e, "Execute stage is busy, cannot advance")
		e.next.Advance(nil, true) // pass a bubble and say we are stalled
		return false
	}
	if e.next.CanAdvance() {
		e.pipeline.sTracef(e, "Advancing to next stage with instruction: %+v\n", e.currentInstruction)
		e.next.Advance(e.currentInstruction, false) // Pass the instruction to the next stage
		e.currentInstruction = i
		return true
	} else {
		e.pipeline.sTracef(e, "Can not advance to %v, CanAdvance returned false", e.next.Name())
		e.next.Advance(nil, false) // pass a bubble and say we are not stalled
		return false
	}
}

func (e *ExecuteStage) Squash() bool {
	e.pipeline.sTracef(e, "Squashing instruction: %+v\n", e.currentInstruction)
	if e.currentInstruction != nil {
		e.pipeline.cpu.unblockIntR(e.currentInstruction.BaseInstruction.Rd)
		e.pipeline.cpu.unblockIntR(e.currentInstruction.BaseInstruction.RMem)
		e.currentInstruction = nil
	}
	e.state = EXEC_free
	return true
}

func (e *ExecuteStage) CanAdvance() bool {
	return e.state == EXEC_free
}

func (e *ExecuteStage) FormatInstruction() string {
	if e.currentInstruction == nil {
		return "<bubble>"
	}
	//format := fmt.Sprintf("%+v", e.currentInstruction)
	return e.currentInstruction.FormatLines()
}