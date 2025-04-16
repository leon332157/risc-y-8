package cpu

import (
	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

func NewCPU(cache *memory.CacheType) *CPU {
	var cpu *CPU
	cpu = &CPU{
		Clock:          0,
		ProgramCounter: 0,
		ALU:            alu.NewALU(),
		Cache:          cache,
		Pipeline: NewPipeline(cpu),
	}
	return cpu;
}
