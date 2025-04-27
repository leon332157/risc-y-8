package cpu

import (
	"fmt"
	"github.com/leon332157/risc-y-8/pkg/types"
)

const (
	DEC_free = iota
	DEC_busy
	DEC_reg_read // busy waiting for register read
)

type DecodeStage struct {
	currInst *InstructionIR
	state    int
	pipe     *Pipeline

	next *ExecuteStage
	prev *FetchStage

	instStr string
}

func signExtend(v int16) uint32 {
	return uint32(int32(v))
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
	d.instStr = "<bubble>"
	return nil
}

func (d *DecodeStage) Name() string {
	return "Decode"
}

func (d *DecodeStage) Execute() {
	if d.currInst == nil {
		d.pipe.sTrace(d, "No current instruction to process, returning early") // For debugging purposes, return early if no instruction is set
		d.instStr = "<bubble>"
		return
	}
	d.state = DEC_busy
	d.currInst.BaseInstruction = new(types.BaseInstruction)      // Create a new BaseInstruction to decode the instruction
	d.currInst.BaseInstruction.Decode(d.currInst.rawInstruction) // Decode the raw instruction into a BaseInstruction
	d.pipe.sTracef(d, "Decoded instruction: %+v\n", d.currInst)  // For debugging purposes
	baseInstruction := d.currInst.BaseInstruction                // For convenience
	//go func() {
		d.instStr = fmt.Sprintf("raw: 0x%08x\n", d.currInst.rawInstruction)
		d.instStr += fmt.Sprintf(
			"OpType: %x\nRd: %x\nDestMemAddr: %x\nRDAux: %x\n",
			baseInstruction.OpType, baseInstruction.Rd, d.currInst.DestMemAddr, d.currInst.RDestAux)
	//}()

	switch baseInstruction.OpType {
	case types.RegImm:

		v, st := d.pipe.cpu.ReadIntR(baseInstruction.Rd)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read register r%v %v", baseInstruction.Rd, st)
			d.state = DEC_busy
			return
		}
		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		d.currInst.Result = v
		d.currInst.Operand = signExtend(baseInstruction.Imm) // sign extend immediate value
		d.currInst.WriteBack = baseInstruction.Rd != 0       // Only write back if Rd is not zero

	case types.RegReg:

		rdv, st := d.pipe.cpu.ReadIntR(baseInstruction.Rd)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read dest register r%v %v", baseInstruction.Rs, st)
			d.state = DEC_busy
			return
		}
		rsv, st := d.pipe.cpu.ReadIntR(baseInstruction.Rs)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read source register r%v %v", baseInstruction.Rs, st)
			d.state = DEC_reg_read
			return
		}
		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		d.pipe.cpu.blockIntR(baseInstruction.Rs)
		d.currInst.Result = rdv
		d.currInst.Operand = rsv
		d.currInst.WriteBack = baseInstruction.Rd != 0 // Only write back if Rd is not zero, otherwise it might be a nop operation

	case types.Control:

		if baseInstruction.RMem == 0 && baseInstruction.Imm == -1 {
			d.pipe.log.Error().Msg("[DecodeStage Execute] Bruh, why are branching to -1? Are you trying to halt?")
			d.pipe.cpu.Halt()
			return
		}
		rmemv, st := d.pipe.cpu.ReadIntR(baseInstruction.RMem)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read control instruction memory source r%v %v", baseInstruction.RMem, st)
			d.state = DEC_reg_read
			return
		}
		d.currInst.DestMemAddr = rmemv
		if baseInstruction.RMem == 0 {
			d.currInst.DestMemAddr = d.pipe.cpu.ProgramCounter // use PC as destination address if RMem is 0
		}
		d.currInst.Operand = signExtend(baseInstruction.Imm) // sign extend immediate value
		d.currInst.WriteBack = true                          // Always write back for control instructions

	case types.LoadStore:

		v, st := d.pipe.cpu.ReadIntR(baseInstruction.Rd)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read load/store instruction destination register r%v %v", baseInstruction.Rd, st)
			d.state = DEC_reg_read
			return
		}
		d.currInst.Result = v
		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		var SP = types.IntegerRegisters["sp"]
		if (baseInstruction.MemMode == types.PUSH) || (baseInstruction.MemMode == types.POP) {
			// If push or pop
			d.currInst.BaseInstruction.RMem = SP
			d.pipe.sTracef(d, "PUSH/POP instruction detected, setting RMem to SP (r%v)", SP)
		}
		// set memory source to be stack pointer so squash can work properlyq
		rmemv, st := d.pipe.cpu.ReadIntR(baseInstruction.RMem) // rmemv should be zero for push pop
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read load/store instruction memory source r%v %v", baseInstruction.RMem, st)
			d.state = DEC_reg_read
			return
		}
		d.pipe.cpu.blockIntR(baseInstruction.RMem)
		d.pipe.sTracef(d, "Read memory source r%v value %v", baseInstruction.RMem, rmemv) // For debugging purposes
		d.currInst.DestMemAddr = rmemv

		d.currInst.Operand = signExtend(d.currInst.BaseInstruction.Imm)
		d.currInst.WriteBack = d.currInst.BaseInstruction.MemMode <= 1 // If LDW or POP
	}
	d.state = DEC_free
	d.pipe.sTracef(d, "Decoded filled instruction: %+v %+v\n", d.currInst, *d.currInst.BaseInstruction)
	//go func() {
		switch baseInstruction.OpType {
		case types.RegReg:
			d.instStr += fmt.Sprintf("Rs: %x\n", baseInstruction.Rs)
			d.instStr += fmt.Sprintf("ALU: %s\n", types.RegALUInverse[baseInstruction.ALU])
		case types.RegImm:
			d.instStr += fmt.Sprintf("ALU: %s\n", types.ImmALUInverse[baseInstruction.ALU])
			d.instStr += fmt.Sprintf("Imm: %x\n", baseInstruction.Imm)
		case types.LoadStore:
			d.instStr += fmt.Sprintf("MemMode: %v\n", baseInstruction.MemMode)
		case types.Control:
			d.instStr += fmt.Sprintf("CtrlMode: %x\n", baseInstruction.CtrlMode)
			d.instStr += fmt.Sprintf("CtrlFlag: %x\n", baseInstruction.CtrlFlag)
		}
		d.instStr += fmt.Sprintf("Result: %x\n", d.currInst.Result)
	//}()
}

// Returns if this stage passed the instruction to the next stage
func (d *DecodeStage) Advance(i *InstructionIR, prevstalled bool) bool {
	if prevstalled {
		d.pipe.sTracef(d, "previous stage %v is stalled", d.prev.Name())
	}
	if d.state != DEC_free {
		d.pipe.sTracef(d, "Decode is busy, cannot advance: %v", d.state)
		d.next.Advance(nil, true) // tell next stage we are stalled, push bubble
		return false
	}
	if d.next.CanAdvance() {
		d.pipe.sTracef(d, "Advancing to next stage with instruction: %+v\n", d.currInst)
		d.next.Advance(d.currInst, false) // Pass the instruction to the next stage
		d.currInst = i                    // take in our next instruction
		return true
	} else {
		d.pipe.sTracef(d, "Can not advance to %v, CanAdvance returned false", d.next.Name())
		d.next.Advance(nil, false) // pass bubble and say we are not stalled
		return false
	}
}

func (d *DecodeStage) Squash() bool {
	d.pipe.sTracef(d, "Squashing instruction: %+v\n", d.currInst) // For debugging purposes
	if d.currInst != nil {
		d.pipe.cpu.unblockIntR(d.currInst.BaseInstruction.Rd)
		d.pipe.cpu.unblockIntR(d.currInst.BaseInstruction.Rs)
		d.pipe.cpu.unblockIntR(d.currInst.BaseInstruction.RMem)
	}
	d.currInst = nil
	d.state = DEC_free
	return true
}

// Returns returns if this stage can take in a new instruction
func (d *DecodeStage) CanAdvance() bool {
	return d.state == DEC_free
}

func (d *DecodeStage) FormatInstruction() string {
	return d.instStr
}
