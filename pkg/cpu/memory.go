package cpu

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/memory"
	"github.com/leon332157/risc-y-8/pkg/types"
)

type MemoryStage struct {
	currInst *InstructionIR  // Pointer to the InstructionIR being processed in this stage
	waiting            bool

	pipeline           *Pipeline       // Reference to the pipeline instance
	next               *WriteBackStage // Next stage in the pipeline
	prev               *ExecuteStage   // Previous stage in the pipeline

	instStr string
}

func (e *MemoryStage) Init(pipeline *Pipeline, next Stage, prev Stage) error {
	if pipeline == nil {
		panic("[Mem Init] pipeline is null")
	}
	e.pipeline = pipeline
	if next == nil {
		panic("[Mem Init] next is null")
	}
	n, ok := next.(*WriteBackStage)
	if !ok {
		return fmt.Errorf("[Mem Init] next stage is not writeback stage")
	}
	if n == nil {
		return fmt.Errorf("[Mem Init] next stage is null")
	}
	e.next = n
	p, ok := prev.(*ExecuteStage)
	if !ok {
		return fmt.Errorf("[Mem Init] prev stage is not execute stage")
	}
	if p == nil {
		return fmt.Errorf("[Mem Init] prev is null")
	}
	e.prev = p
	e.instStr = "<bubble>"
	return nil
}

func (m *MemoryStage) Name() string {
	return "Memory"
}

func (m *MemoryStage) Execute() {
	inst := m.currInst
	if inst == nil {
		m.pipeline.sTrace(m, "No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		m.instStr = "<bubble>"
		return
	}
	m.instStr = fmt.Sprintf("OpType: %x\nMem Mode: %x\nRd: %x\nRMem: %x\nDestMemAddr: %x", inst.BaseInstruction.OpType, inst.BaseInstruction.MemMode, inst.BaseInstruction.Rd, inst.BaseInstruction.RMem, inst.DestMemAddr)
	if inst.BaseInstruction.OpType != types.LoadStore {
		m.pipeline.sTracef(m, "Current instruction is not a load/store type, skipping memory stage execution %+v\n", inst) // For debugging purposes, skip if not a load/store instruction
		return
	} else {
		m.pipeline.sTracef(m, "[MemoryStage Execute] Processing instruction: %+v\n", inst) // For debugging purposes
	}
	cache := m.pipeline.cpu.Cache
	destAddr := uint(inst.DestMemAddr)
	switch inst.BaseInstruction.MemMode {
	case types.LDW, types.POP:
		attempt := cache.Read(destAddr, memory.MEMORY_STAGE) // Attempt to read from cache
		if attempt.State != memory.SUCCESS {
			m.pipeline.sTracef(m, "Failed to load from cache at address 0x%X, state: %s\n", inst.DestMemAddr, memory.LookUpMemoryResult(attempt.State))
		}
		if attempt.State == memory.WAIT || attempt.State == memory.WAIT_NEXT_LEVEL {
			// Handle waiting or next level cache logic here if needed
			m.pipeline.sTracef(m, "Waiting for cache read at address 0x%X\n", inst.DestMemAddr)
			m.waiting = true
			return // Do not proceed further until the cache read is successful
		} else {
			// Successfully read from cache, set the result in the current instruction
			m.currInst.Result = uint32(attempt.Value) // Set the result to the value read from cache
			m.waiting = false
			m.pipeline.sTracef(m, "Successfully loaded from cache at address 0x%X, value: %d\n", inst.DestMemAddr, inst.Result)
			if m.currInst.BaseInstruction.MemMode == types.POP {
				m.currInst.DestMemAddr--
			}
		}

	case types.STW, types.PUSH:
		m.pipeline.sTracef(m, "Attempting to store value %d to cache at address 0x%X\n", m.currInst.Result, inst.DestMemAddr) // For debugging purposes
		writeResult := cache.Write(destAddr, memory.MEMORY_STAGE, m.currInst.Result)                                          // Attempt to write to cache
		if writeResult.State != memory.SUCCESS {
			// Handle failure to write to cache
			m.pipeline.sTracef(m, "Failed to store to cache at address 0x%X, state: %s\n", m.currInst.DestMemAddr, memory.LookUpMemoryResult(writeResult.State))
		}
		if writeResult.State == memory.WAIT || writeResult.State == memory.WAIT_NEXT_LEVEL {
			// Handle waiting or next level cache logic here if needed
			m.pipeline.sTracef(m, "Waiting for cache write at address 0x%X\n", m.currInst.DestMemAddr)
			m.waiting = true
		} else {
			// Successfully wrote to cache
			if m.currInst.BaseInstruction.MemMode == types.PUSH {
				m.currInst.DestMemAddr++
			}
			m.pipeline.sTracef(m, "Successfully stored to cache at address 0x%X\n", m.currInst.DestMemAddr)
			m.pipeline.cpu.unblockIntR(m.currInst.BaseInstruction.Rd) // Unblock the register after successful write
			m.waiting = false // Clear waiting state since the write was successful
		}

	default:
		m.pipeline.log.Panic().Msgf("[Memory Stage] Unsupported memory operation: %d", m.currInst.BaseInstruction.MemMode) // Handle unsupported memory operations
	}
}

func (m *MemoryStage) Advance(i *InstructionIR, prevstalled bool) bool {
	if prevstalled {
		m.pipeline.sTracef(m, "previous stage %v returned stall\n", m.prev.Name())
	}
	if m.waiting {
		m.pipeline.sTrace(m, "still waiting for operation, can not advance\n")
		m.next.Advance(nil, true) // pass a bubble and say we are stalled
		return false
	}
	if m.next.CanAdvance() {
		// Writeback stage should typically always return true for can advance
		m.pipeline.sTracef(m, "Advancing to next stage with instruction: %+v\n", m.currInst)
		m.next.Advance(m.currInst, false) // pass to next stage
		m.currInst = i                    // update my instruction
		return true
	} else {
		m.pipeline.sTracef(m, "Can not advance to %v, CanAdvance returned false", m.next.Name())
		m.next.Advance(nil, false) // pass a bubble and say we are not stalled
	}
	return false
}

func (m *MemoryStage) Squash() bool {
	m.pipeline.sTracef(m, "Squashing instruction: %+v\n", m.currInst) // For debugging purposes
	if m.currInst != nil {
		m.pipeline.cpu.unblockIntR(m.currInst.BaseInstruction.Rd) // Unblock the register if it was blocked
		m.pipeline.cpu.unblockIntR(m.currInst.BaseInstruction.RMem) // Unblock the memory register if it was blocked
		m.pipeline.cpu.unblockIntR(m.currInst.RDestAux) // Unblock the auxiliary register if it was blocked
	}
	m.currInst = nil
	m.waiting = false

	// Cancel request to memory/cache if necessary
	cache := m.pipeline.cpu.Cache
	ram := m.pipeline.cpu.RAM
	// Check if cache/ram currently serving MEM_STAGE
	if cache.Requester() == memory.MEMORY_STAGE {
		m.pipeline.cpu.Cache.CancelRequest()
	}
	if ram.Requester() == memory.MEMORY_STAGE {
		m.pipeline.cpu.RAM.CancelRequest()
	}
	return true
}

func (m *MemoryStage) CanAdvance() bool {
	return (m.next.CanAdvance()) && !m.waiting 
}

func (m *MemoryStage) FormatInstruction() string {
	return m.instStr
}
