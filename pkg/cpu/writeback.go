package cpu

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/types"
)

type WriteBackStage struct {
	currentInstruction *InstructionIR // Pointer to the InstructionIR being processed in this stage

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
	if w.currentInstruction == nil {
		// No instruction to write back, return early
		w.pipeline.sTrace(w, "No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		w.instStr = "<bubble>"
		return
	}
	w.instStr = fmt.Sprintf("OpType: %x\n", w.currentInstruction.BaseInstruction.OpType)
	w.instStr += fmt.Sprintf("Rd: %x\n", w.currentInstruction.BaseInstruction.Rd)
	w.instStr += fmt.Sprintf("Writeback: %v\n", w.currentInstruction.WriteBack)
	w.instStr += fmt.Sprintf("Result: %x\n", w.currentInstruction.Result)
	w.instStr += fmt.Sprintf("RDestAux: %x\n", w.currentInstruction.RDestAux)
	if w.currentInstruction.BaseInstruction.OpType == types.Control {
		w.instStr += fmt.Sprintf("BranchTaken: %v\n", w.currentInstruction.BranchTaken)
		w.instStr += fmt.Sprintf("DestMemAddr: %x\n", w.currentInstruction.DestMemAddr)
		w.instStr += fmt.Sprintf("RMem: %x\n", w.currentInstruction.BaseInstruction.RMem)
		w.instStr += fmt.Sprintf("CtrlModeFlag: %x\n", w.currentInstruction.BaseInstruction.CtrlMode<<4|w.currentInstruction.BaseInstruction.CtrlFlag)
	}
	inst := w.currentInstruction
	w.pipeline.sTracef(w, "Processing instruction: %+v\n", w.currentInstruction) // For debugging purposes
	if w.pipeline.scalarMode {
		w.pipeline.canFetch = true // In scalar mode, we can fetch the next instruction after write back
	}
	if w.currentInstruction.WriteBack {
		if w.currentInstruction.BaseInstruction.OpType == types.Control {
			// Control instruction, write back to the Program Counter and RDestAUX
			w.pipeline.sTrace(w, "Control instruction detected")
			if !inst.BranchTaken {
				w.pipeline.sTrace(w, "Branch not taken, no write back to Program Counter required")
				//w.pipeline.cpu.unblockIntR(w.currentInstruction.BaseInstruction.Rd)
				w.currentInstruction = nil
				return
			}
			if w.currentInstruction.BaseInstruction.Rd == 0 {
				// writing to PC
				w.pipeline.sTracef(w, "Writing to Program Counter directly from control instruction to %v\n", w.currentInstruction.DestMemAddr)
				w.pipeline.cpu.ProgramCounter = w.currentInstruction.DestMemAddr // Update the Program Counter if this is a control instruction
				w.pipeline.SquashALL()
				return
			}
		}
		w.pipeline.sTracef(w, "Writing back result: %v to r%v\n", w.currentInstruction.Result, w.currentInstruction.BaseInstruction.Rd)
		w.pipeline.cpu.unblockIntR(w.currentInstruction.BaseInstruction.Rd)                                         // Unblock the destination register if applicable
		w.pipeline.sTracef(w, "Unblocked register r%v for write back\n", w.currentInstruction.BaseInstruction.Rd)   // For debugging purposes
		_, status := w.pipeline.cpu.WriteIntR(w.currentInstruction.BaseInstruction.Rd, w.currentInstruction.Result) // Write the result to the destination register
		if status != SUCCESS {
			w.pipeline.sTracef(w, "Failed to write back to register r%v: %v\n", w.currentInstruction.BaseInstruction.Rd, status)
			return
		}
		w.pipeline.cpu.WriteIntR(w.currentInstruction.RDestAux, w.currentInstruction.DestMemAddr)
		w.pipeline.sTracef(w, "Written back to rdestaux: %v %v\n", w.currentInstruction.RDestAux, w.currentInstruction.DestMemAddr)
	} else {
		w.pipeline.sTrace(w, "No write-back required for this instruction")
	}
	w.pipeline.sTracef(w, "Write back completed for instruction: %+v\n", w.currentInstruction) // For debugging purposes
	w.currentInstruction = nil
}

func (w *WriteBackStage) Advance(i *InstructionIR, prevstalled bool) bool {
	if prevstalled {
		w.pipeline.sTracef(w, "previous stage %v returned is stalled\n", w.prev.Name())
		return false
	}
	w.pipeline.sTracef(w, "Got with instruction: %+v\n", i) // Debugging output to see which instruction is being processed
	w.currentInstruction = i                                // Set the current instruction to the one passed in, this is used in Execute()

	return true
}

func (w *WriteBackStage) Squash() bool {
	w.pipeline.sTracef(w, "Squashing instruction: %+v\n", w.currentInstruction) // For debugging purposes
	w.currentInstruction = nil
	return true
}

func (w *WriteBackStage) CanAdvance() bool {
	return w.currentInstruction == nil
}

func (w *WriteBackStage) FormatInstruction() string {
	return w.instStr
}
