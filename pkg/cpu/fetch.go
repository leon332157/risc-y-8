package cpu

import (
	"github.com/leon332157/risc-y-8/pkg/memory"
)

type FetchStage struct {
	pipe               *Pipeline // Reference to the pipeline instance
	next               *DecodeStage
	currentInstruction *InstructionIR
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
	return nil
}

func (f *FetchStage) Name() string {
	return "Fetch"
}

func (f *FetchStage) Execute() {
	if f.currentInstruction != nil {
		f.pipe.sTrace(f, ("Currently have an instruction, skipping fetch"))
		return
	}
	f.pipe.sTrace(f, "Fetching instruction from memory")

	cache := f.pipe.cpu.Cache
	read := cache.Read(uint(f.pipe.cpu.ProgramCounter), memory.FETCH_STAGE)
	if read.State != memory.SUCCESS {
		f.pipe.sTraceF(f, "Fetch failed: %v", memory.LookUpMemoryResult(read.State)) // Memory fetch failed
		return
	}

	if read.Value != 0 {
		f.pipe.sTraceF(f, "Fetched instruction: 0x%08x\n", read.Value)
		f.currentInstruction = new(InstructionIR) // Store the fetched instruction
		f.currentInstruction.rawInstruction = read.Value
		f.pipe.cpu.ProgramCounter++
		f.pipe.sTraceF(f, "Increasing ProgramCounter to: %v", f.pipe.cpu.ProgramCounter)
	} else {
		f.pipe.sTrace(f, "Fetched instruction is zero, no valid instruction found")
		return
	}
}

func (f *FetchStage) Advance(_ *InstructionIR, stall bool) bool {
	if f.currentInstruction == nil {
		f.pipe.sTrace(f,"No instruction fetched, cannot advance")
		f.next.Advance(nil, true) // push empty inst and tell next stage we are stalled
		return false
	}
	f.pipe.sTraceF(f, "Advancing to next stage with instruction: 0x%08x", f.currentInstruction.rawInstruction)
	f.next.Advance(f.currentInstruction, false)
	f.currentInstruction = nil // clear out curr inst after advancing
}

func (f *FetchStage) CanAdvance() bool {
	return f.currentInstruction != nil
}