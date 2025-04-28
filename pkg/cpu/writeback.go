package cpu

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/types"
)

type WriteBackStage struct {
	currInst *InstructionIR // Pointer to the InstructionIR being processed in this stage

	pipeline *Pipeline    // Reference to the pipeline instance
	prev     *MemoryStage // Previous stage in the pipeline

	instStr string
}

func (w *WriteBackStage) Init(pipeline *Pipeline, _ Stage, prev Stage) error {
	if pipeline == nil {
		w.instStr = "<bubble>"
		return fmt.Errorf("[WriteBackStage Init] pipeline is null")
	}
	w.pipeline = pipeline
	p, ok := prev.(*MemoryStage)
	if !ok {
		return fmt.Errorf("[fetch Init] prev stage is not memory stage")
	}
	if p == nil {
		return fmt.Errorf("[fetch Init] prev is null")
	}
	w.prev = p
	w.instStr = "<bubble>"
	return nil
}

func (w *WriteBackStage) Name() string {
	return "WriteBack"
}

func (w *WriteBackStage) Execute() {
	if w.currInst == nil {
		// No instruction to write back, return early
		w.pipeline.sTrace(w, "No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		w.instStr = "<bubble>"
		return
	}
	w.instStr = fmt.Sprintf("OpType: %x\n", w.currInst.BaseInstruction.OpType)
	w.instStr += fmt.Sprintf("Rd: %x\n", w.currInst.BaseInstruction.Rd)
	w.instStr += fmt.Sprintf("Result: %x\n", w.currInst.Result)
	w.instStr += fmt.Sprintf("RDestAux: %x\n", w.currInst.RDestAux)
	w.instStr += fmt.Sprintf("ResultAux: %x\n", w.currInst.ResultAux)
	if w.currInst.BaseInstruction.OpType == types.Control {
		w.instStr += fmt.Sprintf("BranchTaken: %v\n", w.currInst.BranchTaken)
		w.instStr += fmt.Sprintf("DestMemAddr: %x\n", w.currInst.DestMemAddr)
		w.instStr += fmt.Sprintf("RMem: %x\n", w.currInst.BaseInstruction.RMem)
		w.instStr += fmt.Sprintf("CtrlModeFlag: %x\n", w.currInst.BaseInstruction.CtrlMode<<4|w.currInst.BaseInstruction.CtrlFlag)
	}

	w.pipeline.sTracef(w, "Processing instruction: %+v\n", w.currInst) // For debugging purposes

	if w.currInst.BaseInstruction.OpType == types.Control {
		// Control instruction, write back to the Program Counter and RDestAUX
		w.pipeline.sTrace(w, "Control instruction detected")
		if w.currInst.BranchTaken {
			if w.currInst.BaseInstruction.Rd == 0 {
				// writing to PC
				w.pipeline.sTracef(w, "Writing to Program Counter directly from control instruction to %v\n", w.currInst.DestMemAddr)
				w.pipeline.cpu.ProgramCounter = w.currInst.DestMemAddr // Update the Program Counter if this is a control instruction
				w.pipeline.SquashALL()
				return
			}
		} else {
			w.pipeline.sTrace(w, "Branch not taken, no write back to Program Counter required")
		}
	}

	w.pipeline.cpu.unblockIntR(w.currInst.BaseInstruction.Rd)
	w.pipeline.sTracef(w, "Unblocked register r%v for write back\n", w.currInst.BaseInstruction.Rd) // For debugging purposes
	w.pipeline.sTracef(w, "Writing back result: %v to r%v\n", w.currInst.Result, w.currInst.BaseInstruction.Rd)
	_, status := w.pipeline.cpu.WriteIntR(w.currInst.BaseInstruction.Rd, w.currInst.Result) // Write the result to the destination register
	if status != SUCCESS {
		w.pipeline.sTracef(w, "Failed to write back to register r%v: %v\n", w.currInst.BaseInstruction.Rd, status)
		return
	}
	w.pipeline.cpu.unblockIntR(w.currInst.RDestAux)
	w.pipeline.sTracef(w, "Unblocked register r%v for write back\n", w.currInst.RDestAux) // For debugging purposes
	w.pipeline.sTracef(w, "Writing back result: %v to r%v\n", w.currInst.ResultAux, w.currInst.RDestAux)
	_, status = w.pipeline.cpu.WriteIntR(w.currInst.RDestAux, w.currInst.ResultAux) // Write the result to the destination register
	if status != SUCCESS {
		w.pipeline.sTracef(w, "Failed to write back to register r%v: %v\n", w.currInst.RDestAux, status)
		return
	}
	w.pipeline.sTracef(w, "Write back completed for instruction: %+v\n", w.currInst) // For debugging purposes
	w.pipeline.sTracef(w, "Write back completed for base instruction: %+v\n", *w.currInst.BaseInstruction) // For debugging purposes
	w.currInst = nil

	if w.pipeline.scalarMode {
		w.pipeline.canFetch = true // In scalar mode, we can fetch the next instruction after write back
	}
}

func (w *WriteBackStage) Advance(i *InstructionIR, prevstalled bool) bool {
	if prevstalled {
		w.pipeline.sTracef(w, "previous stage %v returned is stalled\n", w.prev.Name())
		return false
	}
	w.pipeline.sTracef(w, "Got with instruction: %+v\n", i) // Debugging output to see which instruction is being processed
	w.currInst = i                                          // Set the current instruction to the one passed in, this is used in Execute()

	return true
}

func (w *WriteBackStage) Squash() bool {
	w.pipeline.sTracef(w, "Squashing instruction: %+v\n", w.currInst) // For debugging purposes
	w.pipeline.cpu.unblockIntR(w.currInst.BaseInstruction.Rd)
	w.pipeline.cpu.unblockIntR(w.currInst.BaseInstruction.RMem)
	w.pipeline.cpu.unblockIntR(w.currInst.RDestAux)
	w.currInst = nil
	return true
}

func (w *WriteBackStage) CanAdvance() bool {
	return w.currInst == nil
}

func (w *WriteBackStage) FormatInstruction() string {
	return w.instStr
}
