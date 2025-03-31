package writeback

import (
	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/types"
)

var clock_cycle uint32 = 0

 
func WriteBackStage(regs *alu.Registers, mtowb types.MemToWB) types.WBToMem {

	clock_cycle++

	wbtom := types.WBToMem{}
	
	// check if mtowb is empty and if it is, then return 

	result := mtowb.RegVal

	regs.IntRegisters[mtowb.Reg] = result
	regs.RFlag = mtowb.Flag

	wbtom.Reg = mtowb.Reg
	wbtom.RegVal = result
	
	return wbtom
	
}

