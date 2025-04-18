package cpu

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/memory"
)

type FetchStage struct {
	pipe               *Pipeline // Reference to the pipeline instance
	next               *DecodeStage
	currentInstruction *InstructionIR

	formatCache string
}

func (f *FetchStage) Init(p *Pipeline, next Stage, _ Stage) error {
	if p == nil {
		f.pipe.log.Fatal().Msg("[Fetch Init] pipeline is null")
	}
	f.pipe = p
	n, ok := next.(*DecodeStage)
	if !ok {
		f.pipe.log.Fatal().Msg("[Fetch Init] next stage is not decode stage")
	}
	f.next = n
	f.currentInstruction = nil
	f.formatCache = "<bubble>"
	return nil
}

func (f *FetchStage) Name() string {
	return "Fetch"
}

func (f *FetchStage) Execute() {
	f.formatCache = "<bubble>"
	if f.pipe.scalarMode && f.pipe.canFetch == false {
		f.pipe.sTrace(f, "Cannot fetch instruction right now, write has not completed yet")
		return
	}

	if f.currentInstruction != nil {
		f.pipe.sTracef(f, "Currently have an instruction %+v, skipping fetch", f.currentInstruction)
		return
	}
	//f.pipe.sTrace(f, "Fetching instruction from memory")
	cache := f.pipe.cpu.Cache
	read := cache.Read(uint(f.pipe.cpu.ProgramCounter), memory.FETCH_STAGE)
	if read.State != memory.SUCCESS {
		f.pipe.sTracef(f, "Fetch failed: %v", memory.LookUpMemoryResult(read.State)) // Memory fetch failed
		return
	}

	if read.Value != 0 {
		f.pipe.sTracef(f, "Fetched instruction: 0x%08x\n", read.Value)
		f.currentInstruction = new(InstructionIR) // Store the fetched instruction
		f.currentInstruction.rawInstruction = read.Value
		f.formatCache = fmt.Sprintf("0x%08x\n", f.currentInstruction.rawInstruction)
		f.pipe.cpu.ProgramCounter++
		f.pipe.sTracef(f, "Increasing ProgramCounter to: %v", f.pipe.cpu.ProgramCounter)
		if f.pipe.scalarMode {
			// if in scalar mode, we can only fetch one instruction at a time
			f.pipe.canFetch = false
		}
	} else {
		f.pipe.sTrace(f, "Fetched instruction is zero, no valid instruction found")
		return
	}
}

// Returns if this stage passed the instruction to the next stage
func (f *FetchStage) Advance(_ *InstructionIR, canFetch bool) bool {
	if !canFetch {
		f.pipe.sTrace(f, "Stalling due to cpu stall condition")
	}
	if f.currentInstruction == nil {
		f.pipe.sTrace(f, "No instruction fetched, cannot advance")
		f.next.Advance(nil, true) // pass bubble and say we are stalled
		return false
	}
	if f.next.CanAdvance() {
		f.pipe.sTracef(f, "Advancing to next stage with instruction: 0x%08x", f.currentInstruction.rawInstruction)
		f.next.Advance(f.currentInstruction, false)
		f.currentInstruction = nil // clear out curr inst after advancing
		return true
	} else {
		f.pipe.sTrace(f, "Next stage cannot advance, stalling")
		f.next.Advance(nil, false) // pass bubble and say we are not stalled
		return false
	}
}

func (f *FetchStage) Squash() bool {
	f.pipe.sTracef(f, "Squashing instruction: %+v\n", f.currentInstruction) // For debugging purposes
	f.currentInstruction = nil
	return true
}

// Returns returns if this stage can take in a new instruction
func (f *FetchStage) CanAdvance() bool {
	return f.currentInstruction != nil
}

// returns curr instruction
func (f *FetchStage) FormatInstruction() string {
	return f.formatCache
}
