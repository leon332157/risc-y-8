package writeback

import (
	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/types"
)

func WriteBackStage(regs *alu.Registers, wbsi types.WriteBackStageInput) {

	regs.IntRegisters[wbsi.Reg] = wbsi.RegVal
	regs.RFlag = wbsi.Flag

}