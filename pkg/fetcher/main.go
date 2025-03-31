package fetcher

import (
	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

var next_inst_addr uint32 = 0

func FetchInstruction(mem memory.RAM, regs *alu.Registers) uint32 {

	regs.IntRegisters[0] = next_inst_addr
	inst := mem.Read(int(next_inst_addr))
	next_inst_addr += 1
	return inst

}