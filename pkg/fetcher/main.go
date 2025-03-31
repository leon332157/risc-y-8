package fetcher

import (
	"github.com/leon332157/risc-y-8/pkg/memory"
)

// fetcher check that pc is always +0x10 else squash the pipeline
// should pc be sent as input to each stage?
var next_inst_addr uint32 = 0

func FetchInstruction(mem memory.RAM) uint32 {

	inst := mem.Read(int(next_inst_addr))
	next_inst_addr += 1
	return inst

}