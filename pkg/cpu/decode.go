package cpu

import (
	"github.com/leon332157/risc-y-8/pkg/types"
)

const (
	DEC_free = iota
	DEC_busy
	DEC_reg_read // busy waiting for register read
)

type DecodeStage struct {
	pipe     *Pipeline // Reference to the pipeline instance
	currInst *InstructionIR
	state    int

	next *ExecuteStage
	prev *FetchStage
}

func (d *DecodeStage) Init(pipeline *Pipeline, next Stage, prev Stage) error {
	if pipeline == nil {
		d.pipe.log.Fatal().Msg("[Decode Init] pipeline is null")
	}
	d.pipe = pipeline
	if next == nil {
		d.pipe.log.Fatal().Msg("[Decode Init] next stage is null")
	}
	n, ok := next.(*ExecuteStage)
	if n == nil {
		d.pipe.log.Fatal().Msg("[Decode Init] next stage is null")
	}
	if !ok {
		d.pipe.log.Fatal().Msg("[Decode Init] next stage is not execute stage")
	}

	d.next = n
	p, ok := prev.(*FetchStage)
	if p == nil {
		d.pipe.log.Fatal().Msg("[Decode Init] prev is null")
	}
	if !ok {
		d.pipe.log.Fatal().Msg("[Decode Init] prev stage is not fetch stage")
	}

	d.prev = p
	return nil
}

func (d *DecodeStage) Name() string {
	return "Decode"
}

func (d *DecodeStage) Execute() {
	if d.currInst == nil {
		d.pipe.sTrace(d, "No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		return
	}
	d.state = DEC_busy
	baseInstruction := types.BaseInstruction{} // Create a new BaseInstruction to decode the instruction
	if d.currInst.rawInstruction == 0 {
		d.pipe.log.Panic().Msg("[DecodeStage Execute] currentRawInstruction is zero, cannot decode") // this should not happen
	}
	(&baseInstruction).Decode(d.currInst.rawInstruction) // Decode the raw instruction into a BaseInstruction
	// assume success
	d.currInst.BaseInstruction = baseInstruction // Store the base instruction in the InstructionIR

	switch baseInstruction.OpType {
	case types.RegImm:

		v, st := d.pipe.cpu.ReadIntR(baseInstruction.Rd)
		if st != SUCCESS {
			d.pipe.sTraceF(d, "Failed to read register r%v %v", baseInstruction.Rs, st)
			d.state = DEC_busy
			return
		}
		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		d.currInst.Result = v
		d.currInst.Operand = uint32(baseInstruction.Imm) // sign extend immediate value
		//d.currInst.ALUOp = baseInstruction.ALU
		d.currInst.WriteBack = baseInstruction.Rd != 0 // Only write back if Rd is not zero

	case types.RegReg:

		rdv, st := d.pipe.cpu.ReadIntR(baseInstruction.Rd)
		if st != SUCCESS {
			d.pipe.sTraceF(d, "Failed to read dest register r%v %v", baseInstruction.Rs, st)
			d.state = DEC_busy
			return
		}
		rsv, st := d.pipe.cpu.ReadIntR(baseInstruction.Rs)
		if st != SUCCESS {
			d.pipe.sTraceF(d, "Failed to read source register r%v %v", baseInstruction.Rs, st)
			d.state = DEC_reg_read
			return
		}
		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		d.pipe.cpu.blockIntR(baseInstruction.Rs)
		d.currInst.Result = rdv
		d.currInst.Operand = rsv
		d.currInst.WriteBack = baseInstruction.Rd != 0 // Only write back if Rd is not zero, otherwise it might be a nop operation

	case types.Control:

		if baseInstruction.RMem == 0 {
			d.pipe.sTraceF(d, "Control instruction memory source is 0, pc relative memory address will not be used")
		}
		rmemv, st := d.pipe.cpu.ReadIntR(baseInstruction.RMem)
		if st != SUCCESS {
			d.pipe.sTraceF(d, "Failed to read control instruction memory source r%v %v", baseInstruction.RMem, st)
			d.state = DEC_reg_read
			return
		}
		d.pipe.cpu.blockIntR(baseInstruction.RMem)
		d.currInst.DestMemAddr = rmemv
		d.currInst.Operand = uint32(baseInstruction.Imm) // sign extend immediate value
		//d.currInst.ALUOp = types.IMM_ADD // 0 for add alu operation
		d.currInst.WriteBack = true

	case types.LoadStore:

		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		rmemv, st := d.pipe.cpu.ReadIntR(baseInstruction.RMem)
		if st != SUCCESS {
			d.pipe.sTraceF(d, "Failed to read load/store instruction memory source r%v %v", baseInstruction.RMem, st)
			d.state = DEC_reg_read
			return
		}
		d.pipe.cpu.blockIntR(baseInstruction.RMem)
		d.currInst.DestMemAddr = rmemv
		d.currInst.Operand = uint32(d.currInst.BaseInstruction.Imm)
		d.currInst.WriteBack = d.currInst.BaseInstruction.MemMode <= 1 // If LDW or POP
	}

	d.state = DEC_free
}

func (d *DecodeStage) Advance(i *InstructionIR, stalled bool) bool {
	if stalled {
		d.pipe.sTraceF(d, "previous stage %v is stalled", d.prev.Name())
	}
	if d.state != DEC_free {
		d.pipe.sTraceF(d, "We are busy, cannot advance: %v", d.state)
		d.next.Advance(nil, true) // tell execute stage we are stalled, push empty instruction
		return false
	}
	if d.next.CanAdvance() {
		d.pipe.sTraceF(d, "Advancing to next stage with instruction: %+v\n", d.currInst)
		d.next.Advance(d.currInst, false) // Pass the instruction to the next stage
		d.currInst = i                    // take in our next instruction
	} else {
		d.pipe.sTraceF(d, "Can not advance to %v, CanAdvance returned false", d.next.Name())
	}
}

func (d *DecodeStage) CanAdvance() bool {
	return d.state == DEC_free
}
