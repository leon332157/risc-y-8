package cpu

import (
	"fmt"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

type FetchStage struct {
	pipeline    *Pipeline // Reference to the pipeline instance
	next        *DecodeStage
	currentInstruction *InstructionIR
}

func (f *FetchStage) Init(p *Pipeline, next Stage, _ Stage) error {
	if p == nil {
		return fmt.Errorf("[FetchStage Init] pipeline is null")
	}
	f.pipeline = p
	n,ok:= next.(*DecodeStage) 
	if !ok {
		return fmt.Errorf("[fetch Init] next stage is not decode stage")
	}
	f.next = n
	f.currentInstruction = nil
	return nil
}

func (f *FetchStage) Name() string {
	return "Fetch"
}

func (f *FetchStage) Execute() {
	if f.pipeline == nil {
		panic("[FetchExecute] pipeline pointer is null") // Pipeline not initialized
	}

	// Fetch the instruction from memory using the Program Counter
	cpu := f.pipeline.cpu
	if cpu == nil {
		panic("[FetchExecute] CPU pointer is null") // CPU not initialized
	}

	f.currentInstruction = nil // Reset current instruction before fetching a new one
	
	instruction := cpu.Cache.Read(uint(cpu.ProgramCounter), memory.FETCH_STAGE)
	if instruction.State != memory.SUCCESS {
		fmt.Printf("[FetchExecute] Memory fetch failed: %v", memory.LookUpMemoryResult(instruction.State)) // Memory fetch failed
		return 
	}

	f.currentInstruction = new(InstructionIR) // Store the fetched instruction
	// Increment the Program Counter for the next fetch
	cpu.ProgramCounter++
	f.currentInstruction.rawInstruction = uint32(instruction.Value)
}

func (f *FetchStage) Advance(_ *InstructionIR, stall bool) {
	if f.currentInstruction == nil {
		fmt.Println("FetchStage: No instruction fetched, cannot advance")
		f.next.Advance(nil, true) // Pass 0 instruction to next stage and stall it
		return
	}
	f.next.Advance(f.currentInstruction, false)
}
