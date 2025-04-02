package cpu

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/memory"
	"github.com/leon332157/risc-y-8/pkg/types"
)

const (
	LOAD  = 0b00
	STORE = 0b11
)

type MemoryStage struct {
	pipeline           *Pipeline       // Reference to the pipeline instance
	next               *WriteBackStage // Next stage in the pipeline
	prev               *ExecuteStage   // Previous stage in the pipeline
	currentInstruction *InstructionIR  // Pointer to the InstructionIR being processed in this stage
	waiting            bool
}

func (e *MemoryStage) Init(pipeline *Pipeline, next Stage, prev Stage) error {
	if pipeline == nil {
		panic("[DecodeStage Init] pipeline is null")
	}
	e.pipeline = pipeline
	if next == nil {
		panic("[DecodeStage Init] next is null")
	}
	n, ok := next.(*WriteBackStage)
	if !ok {
		return fmt.Errorf("[fetch Init] next stage is not writeback stage")
	}
	if n == nil {
		return fmt.Errorf("[fetch Init] next stage is null")
	}
	e.next = n
	p, ok := prev.(*ExecuteStage)
	if !ok {
		return fmt.Errorf("[fetch Init] prev stage is not execute stage")
	}
	if p == nil {
		return fmt.Errorf("[fetch Init] prev is null")
	}
	e.prev = p
	return nil
}

func (m *MemoryStage) Name() string {
	return "Memory"
}

func (m *MemoryStage) Execute() {
	if m.currentInstruction == nil {
		fmt.Println("[MemoryStage Execute] No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		return
	}
	if m.currentInstruction.BaseInstruction.OpType != types.LoadStore {
		fmt.Printf("[MemoryStage Execute] Current instruction is not a load/store type, skipping memory stage execution %+v\n", m.currentInstruction) // For debugging purposes, skip if not a load/store instruction
		return
	} else {
		fmt.Printf("[MemoryStage Execute] Processing instruction: %+v\n", m.currentInstruction) // For debugging purposes
	}
	cache := m.pipeline.cpu.Cache
	destAddr := uint(m.currentInstruction.DestMemAddr)
	switch m.currentInstruction.MemOp {
	case LOAD:
		attempt := cache.Read(destAddr, memory.MEMORY_STAGE) // Attempt to read from cache
		if attempt.State != memory.SUCCESS {
			fmt.Printf("[Memory Stage] Failed to load from cache at address 0x%X, state: %s\n", m.currentInstruction.DestMemAddr, memory.LookUpMemoryResult(attempt.State))
		}
		if attempt.State == memory.WAIT || attempt.State == memory.WAIT_NEXT_LEVEL {
			// Handle waiting or next level cache logic here if needed
			fmt.Printf("[Memory Stage] Waiting for cache read at address 0x%X\n", m.currentInstruction.DestMemAddr)
			m.waiting = true
			return // Do not proceed further until the cache read is successful
		} else {
			// Successfully read from cache, set the result in the current instruction
			m.currentInstruction.Result = uint32(attempt.Value) // Set the result to the value read from cache
			m.waiting = false
			fmt.Printf("[Memory Stage] Successfully loaded from cache at address 0x%X, value: %d\n", m.currentInstruction.DestMemAddr, m.currentInstruction.Result)
		}
	case STORE:
		writeResult := cache.Write(destAddr, memory.MEMORY_STAGE, m.currentInstruction.Result) // Attempt to write to cache
		if writeResult.State != memory.SUCCESS {
			// Handle failure to write to cache
			fmt.Printf("[Memory Stage] Failed to store to cache at address 0x%X, state: %s\n", m.currentInstruction.DestMemAddr, memory.LookUpMemoryResult(writeResult.State))
			if writeResult.State == memory.WAIT || writeResult.State == memory.WAIT_NEXT_LEVEL {
				// Handle waiting or next level cache logic here if needed
				fmt.Printf("[Memory Stage] Waiting for cache write at address 0x%X\n", m.currentInstruction.DestMemAddr)
				m.waiting = true
			} else {
				// Successfully wrote to cache
				fmt.Printf("[Memory Stage] Successfully stored to cache at address 0x%X\n", m.currentInstruction.DestMemAddr)
				m.waiting = false // Clear waiting state since the write was successful
			}
		}
	default:
		panic(fmt.Sprintf("[Memory Stage] Unsupported memory operation: %d", m.currentInstruction.MemOp)) // Handle unsupported memory operations
	}
}

func (m *MemoryStage) Advance(i *InstructionIR, stalled bool) {
	if stalled {
		fmt.Printf("[%v] previous stage %v returned stall\n", m.Name(), m.prev.Name())
		//m.next.Advance(nil, true)
		//return
	}
	if m.waiting {
		fmt.Printf("[%v] still waiting for operation, can not advance\n", m.Name())
		m.next.Advance(nil, true)
		return
	}
	fmt.Printf("[%v] Advancing to next stage with instruction: %+v\n", m.Name(), m.currentInstruction)
	m.next.Advance(m.currentInstruction, false)
	m.currentInstruction = i
}
