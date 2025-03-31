package fetcher

import (
	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/memory"
	"github.com/leon332157/risc-y-8/pkg/types"
)

func FetchStageToDecode(mem memory.RAM, regs *alu.Registers) types.FetchToDecode {

	inst := types.FetchToDecode{}
	pc := regs.IntRegisters[0]

	regs.IntRegisters[0] = pc
	inst.MemInst = mem.Read(int(pc))
	pc++

	return inst

}