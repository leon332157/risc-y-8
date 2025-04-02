package cpu

import (
	"fmt"
	"github.com/leon332157/risc-y-8/pkg/types"
)

type WriteBackStage struct {
	pipeline           *Pipeline      // Reference to the pipeline instance
	prev               *MemoryStage   // Previous stage in the pipeline
	currentInstruction *InstructionIR // Pointer to the InstructionIR being processed in this stage
	//currentRawInstruction uint32 // For tracking the current instruction being processed, if needed
}

func (w *WriteBackStage) Init(pipeline *Pipeline, _ Stage, prev Stage) error {
	if pipeline == nil {
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
	return nil
}

func (w *WriteBackStage) Name() string {
	return "WriteBack"
}

func (w *WriteBackStage) Execute() {
	if w.currentInstruction == nil {
		// No instruction to write back, return early
		fmt.Println("[WriteBackStage Execute] No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		return
	}
	// Perform the write-back operation, typically writing the result back to the register file or memory
	// TODO: Need to check for writing enable?
	if w.currentInstruction.WriteBack {
		if w.currentInstruction.BaseInstruction.OpType == types.Control {
			fmt.Println("BRANCH")
			if w.currentInstruction.RDestAux != 0 {
				// Write to auxiliary register
				w.pipeline.cpu.IntRegisters[w.currentInstruction.RDestAux].Value = w.pipeline.cpu.ProgramCounter // ?? +1
			}
			if w.currentInstruction.BaseInstruction.Rd == 0 {
				// writing to PC
				fmt.Printf("[WriteBackStage] Writing to Program Counter directly from control instruction to %v\n", w.currentInstruction.DestMemAddr)
				w.pipeline.SquashALL()
				w.pipeline.cpu.ProgramCounter = w.currentInstruction.DestMemAddr // Update the Program Counter if this is a control instruction
			}
		}
		fmt.Printf("[WriteBackStage] Writing back result: %v to r%v\n", w.currentInstruction.Result, w.currentInstruction.BaseInstruction.Rd)
		w.pipeline.cpu.unblockRegister(w.currentInstruction.BaseInstruction.Rd)                                     // Unblock the destination register if applicable
		(&w.pipeline.cpu.IntRegisters[w.currentInstruction.BaseInstruction.Rd]).Value = w.currentInstruction.Result // Write the result back to the register file
	} else {
		fmt.Println("[WriteBackStage] No write-back required for this instruction")
	}
	w.currentInstruction = nil
}

func (w *WriteBackStage) Advance(i *InstructionIR, stalled bool) {
	if stalled {
		fmt.Printf("[%v] previous stage %v returned stall\n", w.Name(), w.prev.Name())
		//return
	}
	fmt.Printf("[WriteBackStage] Got with instruction: %+v\n", i) // Debugging output to see which instruction is being processed
	w.currentInstruction = i // Set the current instruction to the one passed in, this is used in Execute()
}
