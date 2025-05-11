package cpu

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/memory"
)

type FetchStage struct {
	currInst *InstructionIR

	pipe *Pipeline
	next *DecodeStage

	InstStr string
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
	f.currInst = nil
	f.InstStr = "<bubble>"
	return nil
}

func (f *FetchStage) Name() string {
	return "Fetch"
}

func (f *FetchStage) Execute() {
	if f.currInst != nil {
		f.InstStr = fmt.Sprintf("<blocked>\nraw: 0x%08x\n", f.currInst.rawInstruction)
	} else {
		f.InstStr = "<bubble>"
	}
	if f.pipe.scalarMode && f.pipe.canFetch == false {
		f.pipe.sTrace(f, "Cannot fetch instruction right now, writeback has not completed yet")
		f.InstStr = "Waiting for writeback . . ."
		return
	}
	if f.currInst != nil {
		f.pipe.sTracef(f, "Currently have an instruction %+v, skipping fetch", f.currInst)
		//f.InstStr = "Skipped"
		return
	}
	f.InstStr = "Fetching \ninstruction . . ."
	cache := f.pipe.cpu.Cache
	read := cache.Read(uint(f.pipe.cpu.ProgramCounter), memory.FETCH_STAGE)
	if read.State != memory.SUCCESS {
		f.pipe.sTracef(f, "Fetch failed: %v", memory.LookUpMemoryResult(read.State)) // Memory fetch failed
		return
	}

	if read.Value != 0 {
		f.pipe.sTracef(f, "Fetched instruction: 0x%08x\n", read.Value)
		f.InstStr = fmt.Sprintf("Fetched instruction: 0x%08x\n", read.Value)
		f.currInst = new(InstructionIR) // Store the fetched instruction
		f.currInst.rawInstruction = read.Value
		f.InstStr = fmt.Sprintf("raw: 0x%08x\n", f.currInst.rawInstruction)
		f.pipe.cpu.ProgramCounter++
		f.pipe.sTracef(f, "Increasing ProgramCounter to: %v", f.pipe.cpu.ProgramCounter)
		if f.pipe.scalarMode {
			// if in scalar mode, we can only fetch one instruction at a time
			f.pipe.canFetch = false
		}
	} else {
		f.pipe.sTrace(f, "Fetched instruction is zero, no valid instruction found")
		f.InstStr = "raw: 0x0\n"
		f.pipe.cpu.Halt()
		return
	}
}

// Returns if this stage passed the instruction to the next stage
func (f *FetchStage) Advance(_ *InstructionIR, canFetch bool) bool {
	if !canFetch {
		f.pipe.sTrace(f, "Stalling due to cpu stall condition")
	}
	if f.currInst == nil {
		f.pipe.sTrace(f, "No instruction fetched, cannot advance")
		f.next.Advance(nil, true) // pass bubble and say we are stalled
		return false
	}
	if f.next.CanAdvance() {
		f.pipe.sTracef(f, "Advancing to next stage with instruction: 0x%08x", f.currInst.rawInstruction)
		f.next.Advance(f.currInst, false)
		f.currInst = nil // clear out curr inst after advancing
		return true
	} else {
		f.pipe.sTrace(f, "Next stage cannot advance, stalling")
		f.next.Advance(nil, false) // pass bubble and say we are not stalled
		return false
	}
}

func (f *FetchStage) Squash() bool {
	f.pipe.sTracef(f, "Squashing instruction: %+v\n", f.currInst) // For debugging purposes
	f.currInst = nil

	// Cancel request to memory/cache if necessary
	cache := f.pipe.cpu.Cache
	ram := f.pipe.cpu.RAM
	// Check if cache/ram currently serving FETCH
	if cache.Requester() == memory.FETCH_STAGE {
		f.pipe.sTrace(f, "Cancelling Cache Fetch Request") // for debugging
		f.pipe.cpu.Cache.CancelRequest()
	}
	if ram.Requester() == memory.FETCH_STAGE {
		f.pipe.sTrace(f, "Cancelling RAM Fetch Request") // for debugging
		f.pipe.cpu.RAM.CancelRequest()
	}

	return true
}

// Returns returns if this stage can take in a new instruction
func (f *FetchStage) CanAdvance() bool {
	return f.currInst != nil
}

// returns current instruction formatted
func (f *FetchStage) FormatInstruction() string {
	return f.InstStr
}
