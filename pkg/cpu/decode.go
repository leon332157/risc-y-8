package cpu

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/types"
)

type decodeState int

const (
	DEC_free decodeState = iota
	DEC_base_decoded
	DEC_decoded
	DEC_reg_read
)

func LookUpStateDec(s decodeState) string {
	switch s {
	case DEC_free:
		return "DEC_free"
	case DEC_base_decoded:
		return "DEC_base_decoded"
	case DEC_decoded:
		return "DEC_decoded"
	case DEC_reg_read:
		return "DEC_reg_read"
	default:
		return "UNKNOWN"
	}
}

type DecodeStage struct {
	currInst *InstructionIR
	state    decodeState
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
	if d.state < DEC_base_decoded {
		d.currInst.BaseInstruction = new(types.BaseInstruction)      // Create a new BaseInstruction to decode the instruction
		d.currInst.BaseInstruction.Decode(d.currInst.rawInstruction) // Decode the raw instruction into a BaseInstruction
		d.state = DEC_base_decoded
	} else {
		d.pipe.sTrace(d, "Already decoded base instruction, skipping decode")
	}
	d.pipe.sTracef(d, "Decoded instruction: %+v", d.currInst)
	d.pipe.sTracef(d, "Decoded instruction base: %+v", *d.currInst.BaseInstruction)
	baseInstruction := d.currInst.BaseInstruction // For convenience
	//go func() {
	d.instStr = fmt.Sprintf(
		"statebefore:%v\nraw: 0x%08x\nOpType: %x\nRd: %x\nRDAux: %x\n",
		LookUpStateDec(d.state),
		d.currInst.rawInstruction,
		baseInstruction.OpType,
		baseInstruction.Rd,
		d.currInst.RDestAux)
	//}()
	if d.state == DEC_decoded{
		d.pipe.sTracef(d, "Already decoded instruction, skipping decode")
		return
	}
	switch baseInstruction.OpType {
	case types.RegImm:

		v, st := d.pipe.cpu.ReadIntR(baseInstruction.Rd)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read register r%v %v", baseInstruction.Rd, st)
			d.state = DEC_reg_read
			return
		} 
		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		d.currInst.Result = v
		d.currInst.Operand = signExtend(baseInstruction.Imm) // sign extend immediate value
		d.state = DEC_decoded

	case types.RegReg:

		rdv, st := d.pipe.cpu.ReadIntR(baseInstruction.Rd)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read dest register r%v %v", baseInstruction.Rs, st)
			d.state = DEC_reg_read
			return
		}
		rsv, st := d.pipe.cpu.ReadIntR(baseInstruction.Rs)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read source register r%v %v", baseInstruction.Rs, st)
			d.state = DEC_reg_read
			return
		}
		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		d.currInst.Result = rdv
		d.currInst.Operand = rsv
		d.state = DEC_decoded

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
		d.state = DEC_decoded

	case types.LoadStore:

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
		d.pipe.sTracef(d, "Read memory source r%v value %v", baseInstruction.RMem, rmemv) // For debugging purposes
		d.currInst.DestMemAddr = rmemv

		d.currInst.Operand = signExtend(d.currInst.BaseInstruction.Imm)

		v, st := d.pipe.cpu.ReadIntR(baseInstruction.Rd)
		if st != SUCCESS {
			d.pipe.sTracef(d, "Failed to read load/store instruction destination register r%v %v", baseInstruction.Rd, st)
			d.state = DEC_reg_read
			return
		}
		d.currInst.Result = v
		d.pipe.cpu.blockIntR(baseInstruction.RMem)
		d.pipe.cpu.blockIntR(baseInstruction.Rd)
		d.state = DEC_decoded
	}
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
		d.instStr += fmt.Sprintf("MemMode: %v\nDestMem: %v\n", baseInstruction.MemMode, d.currInst.DestMemAddr)
	case types.Control:
		d.instStr += fmt.Sprintf("CtrlMode: %x\n", baseInstruction.CtrlMode)
		d.instStr += fmt.Sprintf("CtrlFlag: %x\n", baseInstruction.CtrlFlag)
		d.instStr += fmt.Sprintf("DestMem: %v",d.currInst.DestMemAddr)
	}
	d.instStr += fmt.Sprintf("Result: %x\n", d.currInst.Result)
	d.instStr += fmt.Sprintf("state after dec: %v", LookUpStateDec(d.state))
	//}()
}

// Returns if this stage passed the instruction to the next stage
func (d *DecodeStage) Advance(i *InstructionIR, prevstalled bool) bool {
	if prevstalled {
		d.pipe.sTracef(d, "previous stage %v is stalled", d.prev.Name())
	}
	if d.next.CanAdvance() {
		if d.state > DEC_decoded {
			// if state is above 0, meaning that decode is busy with it's own work
			d.pipe.sTracef(d, "Decode is busy, cannot advance: %v", d.state)
			d.next.Advance(nil, true) // tell next stage we are stalled, push bubble
			return false
		}
		d.pipe.sTracef(d, "Advancing to next stage with instruction: %+v\n", d.currInst)
		d.next.Advance(d.currInst, false) // Pass the instruction to the next stage
		d.state = DEC_free
		d.currInst = i // take in our next instruction
		return true
	} else {
		d.pipe.sTracef(d, "Can not advance to %v, CanAdvance returned false", d.next.Name())
		d.next.Advance(nil, false) // pass bubble and say we are not stalled
		return false
	}
}

func (d *DecodeStage) Squash() bool {
	d.pipe.sTracef(d, "Squashing instruction: %+v\n", d.currInst) // For debugging purposes
	if d.currInst != nil && d.currInst.BaseInstruction != nil {
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
	return d.state < DEC_reg_read && d.next.CanAdvance()
}

func (d *DecodeStage) FormatInstruction() string {
	return d.instStr
}
