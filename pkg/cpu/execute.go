package cpu

import (
	"fmt"
	"math/bits"

	"github.com/leon332157/risc-y-8/pkg/types"
)

type ExecuteStage struct {
	currInst   *InstructionIR
	state      ExecState
	cyclesLeft uint
	pipeline   *Pipeline // Reference to the pipeline instance

	next *MemoryStage // Next stage in the pipeline
	prev *DecodeStage // Previous stage in the pipeline

	instStr string
}

type ExecState int

const (
	EXEC_free ExecState = iota
	EXEC_done
	EXEC_busy_int    // busy waiting for integer alu to finish
	EXEC_busy_float  // busy waiting for fpu to finish
	EXEC_busy_vector // busy waiting for vector unit to finish
)

func lookUpStateExec(s ExecState) string {
	switch s {
	case EXEC_free:
		return "free"
	case EXEC_done:
		return "done"
	case EXEC_busy_int:
		return "busy_int"
	case EXEC_busy_float:
		return "busy_float"
	case EXEC_busy_vector:
		return "busy_vector"
	default:
		return "unknown state"
	}
}

const DIV_DELAY = 4

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
	e.instStr = "<bubble>"
	return nil
}

func (e *ExecuteStage) Name() string {
	return "Execute"
}

func (e *ExecuteStage) ALURI() {
	// Perform ALU operation for RegImm type instruction
	if e.state == EXEC_free && e.cyclesLeft == 0 {
		e.pipeline.sTrace(e, "adding at least 1 cycle delay")
		e.cyclesLeft = 1
		e.state = EXEC_busy_int
		return
	}
	if e.state == EXEC_busy_int && e.cyclesLeft > 0 {
		e.pipeline.sTrace(e, "busy waiting for integer alu to finish")
		e.cyclesLeft--
		if e.cyclesLeft > 0 {
			return
		}
	}
	if e.state == EXEC_busy_int && e.cyclesLeft == 0 {
		e.pipeline.sTrace(e, "integer alu finished")
		inst := e.currInst
		op1 := inst.Result  // First operand (from a register)
		op2 := inst.Operand // Second operand (immediate value)
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
			inst.Result = uint32(op2 & 0xFFFF)
		case types.IMM_LDX:
			inst.Result = op2
		case types.IMM_CMP:
			e.pipeline.cpu.unblockIntR(inst.BaseInstruction.Rd)
			inst.BaseInstruction.Rd = 0 // Set Rd to 0 for comparison operations
			e.pipeline.cpu.ALU.Sub(op1, op2)
		default:
			e.pipeline.log.Panic().Msg("unsupported ALU operation for RegImm type instruction") // Handle unsupported ALU operations
		}

		e.pipeline.sTracef(e, "ALURI operation result: %v", inst.Result) // For debugging purposes, log the result of the ALU operation
		e.instStr += fmt.Sprintf("ALU: %v\nRd: %x\nResult: %x", types.ImmALUInverse[e.currInst.BaseInstruction.ALU], e.currInst.BaseInstruction.Rd, e.currInst.Result)
		e.state = EXEC_done
		return
	}
	panic("invalid state RI")
}

func (e *ExecuteStage) ALURR() {
	if e.state == EXEC_free && e.cyclesLeft == 0 {
		e.pipeline.sTrace(e, "adding at least 1 cycle delay")
		e.cyclesLeft = 1
		if e.currInst.BaseInstruction.ALU == types.REG_DIV || e.currInst.BaseInstruction.ALU == types.REG_REM {
			e.cyclesLeft = DIV_DELAY
		}
		e.state = EXEC_busy_int
		return
	}
	if e.state == EXEC_busy_int && e.cyclesLeft > 0 {
		e.pipeline.sTrace(e, "busy waiting for integer alu to finish")
		e.cyclesLeft--
		if e.cyclesLeft > 0 {
			return
		}
	}
	if e.state == EXEC_busy_int && e.cyclesLeft == 0 {
		inst := e.currInst
		op1 := inst.Result
		op2 := inst.Operand
		switch inst.BaseInstruction.ALU {
		case types.REG_ADD:
			inst.Result = e.pipeline.cpu.ALU.Add(op1, op2)
		case types.REG_SUB:
			inst.Result = e.pipeline.cpu.ALU.Sub(op1, op2)
		case types.REG_MUL:
			inst.Result = e.pipeline.cpu.ALU.Mul(op1, op2)
		case types.REG_DIV:
			if op2 == 0 {
				panic("division by zero")
			}
			if e.state == EXEC_free && e.cyclesLeft == 1 {
				e.cyclesLeft = DIV_DELAY
			}
			if e.cyclesLeft == 0 {
				e.pipeline.sTrace(e, "integer division finished")
				inst.Result = e.pipeline.cpu.ALU.Div(op1, op2)
			}
		case types.REG_REM:
			if op2 == 0 {
				panic("division by zero on remainder")
			}
			if e.state == EXEC_free && e.cyclesLeft == 1 {
				e.cyclesLeft = DIV_DELAY
			}
			if e.cyclesLeft == 0 {
				e.pipeline.sTrace(e, "integer remainder finished")
				inst.Result = e.pipeline.cpu.ALU.Rem(op1, op2)
			}
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
			e.pipeline.cpu.ALU.Sub(op1, op2)
		case types.REG_CPY:
			inst.Result = op2
		case types.REG_NSA:
			inst.Result = uint32(bits.OnesCount32(op2))
		default:
			e.pipeline.log.Panic().Msg("unsupported ALU operation for RegReg type instruction") // Handle unsupported ALU operations
		}

		e.pipeline.sTracef(e, "ALURR operation result: %v", inst.Result) // For debugging purposes, log the result of the ALU operation
		e.instStr += fmt.Sprintf("ALU: %v\nRd: %x\nResult: %x", types.RegALUInverse[e.currInst.BaseInstruction.ALU], e.currInst.BaseInstruction.Rd, e.currInst.Result)
		e.state = EXEC_done
		return
	}
	panic("invalid state RR")
}

func (e *ExecuteStage) calculateMemAddr(base uint32, displacement int32) uint32 {
	res := (int32(base) + displacement) % int32(e.pipeline.cpu.RAM.SizeWords()) // Calculate the memory address for load/store instructions based on the operands
	e.pipeline.sTracef(e, "calculated memory address: %v", res)                 // For debugging purposes, log the calculated memory address
	if res < 0 {
		panic("negative memory address")
	}
	return uint32(res)
}

func (e *ExecuteStage) LoadStore() {
	if e.state == EXEC_free && e.cyclesLeft == 0 {
		e.pipeline.sTrace(e, "adding at least 1 cycle delay")
		e.cyclesLeft = 1
		e.state = EXEC_busy_int
		return
	}
	if e.state == EXEC_busy_int && e.cyclesLeft > 0 {
		e.pipeline.sTrace(e, "busy waiting for integer alu to finish")
		e.cyclesLeft--
		if e.cyclesLeft > 0 {
			return
		}
	}
	if e.state == EXEC_busy_int && e.cyclesLeft == 0 {
		inst := e.currInst
		switch inst.BaseInstruction.MemMode {
		case types.LDW:
			e.pipeline.sTracef(e, "ldw, calculating memory addr from %v + %v", inst.DestMemAddr, int32(inst.Operand))
			inst.DestMemAddr = e.calculateMemAddr(inst.DestMemAddr, int32(inst.Operand))
		case types.POP, types.PUSH:
			e.pipeline.sTrace(e, "pop/push")
			inst.RDestAux = types.IntegerRegisters["sp"]
		case types.STW:
			e.pipeline.sTracef(e, "stw, calculating memory addr from %v + %v", inst.DestMemAddr, int32(inst.Operand))
			inst.DestMemAddr = e.calculateMemAddr(inst.DestMemAddr, int32(inst.Operand))
		default:
			e.pipeline.log.Panic().Msg("unsupported memory operation for LoadStore type instruction")
		}
		if inst.BaseInstruction.MemMode == types.PUSH {
			inst.ResultAux = inst.DestMemAddr + 1
		}
		if inst.BaseInstruction.MemMode == types.POP {
			inst.DestMemAddr--
			inst.ResultAux = inst.DestMemAddr
		}
		e.instStr += fmt.Sprintf("MemMode: %v\nRd: %x\nRMem: %x\nDestMemAddr: %x\nRdAux %x\nAuxVal: %x", inst.BaseInstruction.MemMode, inst.BaseInstruction.Rd, inst.BaseInstruction.RMem, inst.DestMemAddr, inst.RDestAux, inst.ResultAux)
		e.state = EXEC_done
		return
	}
	panic("invalid state LS")
}

func combineFlags(ctrlMode, ctrlFlag uint8) uint8 {
	return ctrlMode<<4 | ctrlFlag
}

func (e *ExecuteStage) Control() {
	if e.state == EXEC_free && e.cyclesLeft == 0 {
		e.pipeline.sTrace(e, "adding at least 1 cycle delay")
		e.cyclesLeft = 1
		e.state = EXEC_busy_int
		return
	}
	if e.state == EXEC_busy_int && e.cyclesLeft > 0 {
		e.pipeline.sTrace(e, "busy waiting for integer alu to finish")
		e.cyclesLeft--
		if e.cyclesLeft > 0 {
			return
		}
	}
	if e.state == EXEC_busy_int && e.cyclesLeft == 0 {
		inst := e.currInst
		combiFlag := combineFlags(inst.BaseInstruction.CtrlMode, inst.BaseInstruction.CtrlFlag)
		inst.DestMemAddr = e.calculateMemAddr(inst.DestMemAddr, int32(inst.Operand))
		alu := e.pipeline.cpu.ALU
		switch combiFlag {
		case types.GetModeFlag(types.UNC): // unconditional branch
			inst.BranchTaken = true
		case types.GetModeFlag(types.CALL): // call
			inst.BranchTaken = true
			inst.RDestAux = types.IntegerRegisters["lr"]
			inst.ResultAux = e.pipeline.cpu.ProgramCounter
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
		e.instStr += fmt.Sprintf("CtrlMode: %x\nCtrlFlag: %x\nDestMemAddr: %x\nRDestAux: %x\nAuxVal: %x\nBranchTaken: %v\n", inst.BaseInstruction.CtrlMode, inst.BaseInstruction.CtrlFlag, inst.DestMemAddr, inst.RDestAux, inst.ResultAux, inst.BranchTaken)
		e.state = EXEC_done
		return
	}
	panic("invalid state Control")
}

func (e *ExecuteStage) Execute() {
	if e.currInst == nil {
		// If there is no instructionIR to execute, return early
		e.pipeline.sTrace(e, "No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		e.instStr = "<bubble>"
		return
	}
	e.instStr = fmt.Sprintf("State: %v\n", lookUpStateExec(e.state))
	e.pipeline.sTracef(e, "Executing instruction: %+v\n", e.currInst)
	e.pipeline.sTracef(e, "Executing instruction base: %+v\n", *e.currInst.BaseInstruction)
	e.instStr += fmt.Sprintf("OpType: %#v\nCyl Left before: %v\n", types.LookUpOpType(e.currInst.BaseInstruction.OpType), e.cyclesLeft)
	if e.state < EXEC_free {
		panic("nope")
	}
	if e.state != EXEC_done {
		e.pipeline.sTrace(e, "free, executing")
		switch e.currInst.BaseInstruction.OpType {
		case types.RegImm:
			e.ALURI()
		case types.RegReg:
			e.ALURR()
		case types.LoadStore:
			e.LoadStore()
		case types.Control:
			e.Control()
		default:
			panic("unsupported instruction type in Execute stage") // Handle unsupported instruction types
		}
		e.instStr += fmt.Sprintf("\nCyl Left after: %v\n", e.cyclesLeft)
	} else {
		e.pipeline.sTrace(e, "Already executed instruction, not executing")
	}
}

func (e *ExecuteStage) Advance(i *InstructionIR, prevstalled bool) bool {
	if prevstalled {
		e.pipeline.sTracef(e, "previous stage %v returned stall\n", e.prev.Name())
	}
	if e.next.CanAdvance() {
		if e.state > EXEC_done {
			e.pipeline.sTracef(e, "Execute stage is busy, cannot advance: %v", e.state)
			e.next.Advance(nil, true) // tell next stage we are stalled
			return false
		}
		e.pipeline.sTracef(e, "Advancing to next stage with instruction: %+v\n", e.currInst)
		e.next.Advance(e.currInst, false) // Pass the instruction to the next stage
		e.state = EXEC_free
		e.currInst = i
		return true
	} else {
		e.pipeline.sTracef(e, "Can not advance to %v, CanAdvance returned false", e.next.Name())
		e.next.Advance(nil, false) // pass a bubble and say we are not stalled
		return false
	}
}

func (e *ExecuteStage) Squash() bool {
	e.pipeline.sTracef(e, "Squashing instruction: %+v\n", e.currInst)
	if e.currInst != nil {
		e.pipeline.cpu.unblockIntR(e.currInst.BaseInstruction.Rd)
		e.pipeline.cpu.unblockIntR(e.currInst.BaseInstruction.RMem)
		e.currInst = nil
	}
	e.state = EXEC_free
	e.cyclesLeft = 0
	return true
}

func (e *ExecuteStage) CanAdvance() bool {

	return e.next.CanAdvance() && e.state <= EXEC_done
}

func (e *ExecuteStage) FormatInstruction() string {
	return e.instStr
}
