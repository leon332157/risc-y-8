package simulator

import (
	"fmt"

	CPUpkg "github.com/leon332157/risc-y-8/pkg/cpu"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

type System struct {
	CPU   *CPUpkg.CPU
	RAM   *memory.RAM
	Cache *memory.CacheType
}

type readStateHook func(sys *System) bool

func NewSystem(initRamContent []uint32, disableCache, disablePipeline bool) System {
	sys := System{}
	ram := memory.CreateRAM(100, 8, 10)
	sys.RAM = &ram
	sys.CPU = new(CPUpkg.CPU)
	copy(sys.RAM.Contents, initRamContent)
	cache := memory.CreateCacheDefault(sys.RAM) // Create a cache with default settings, 8 sets, 2 ways, no delay                  // Create a new ALU instance
	if disableCache {
		cache = memory.CreateCache(8, 2, 1, 0, sys.RAM) // make one word per line if cache disable
	}
	sys.Cache = &cache
	pipeline := CPUpkg.NewPipeline(sys.CPU, disablePipeline) // scalar is false
	sys.CPU.Init(sys.Cache, sys.RAM, pipeline, nil)          // Initialize the CPU with the cache and no pipeline yet
	fs := new(CPUpkg.FetchStage)
	ds := new(CPUpkg.DecodeStage)
	es := new(CPUpkg.ExecuteStage)
	ms := new(CPUpkg.MemoryStage)
	ws := new(CPUpkg.WriteBackStage)
	fs.Init(pipeline, ds, nil)
	ds.Init(pipeline, es, fs)
	es.Init(pipeline, ms, ds)
	ms.Init(pipeline, ws, es)
	ws.Init(pipeline, nil, ms)
	pipeline.AddStages(ws, ms, es, ds, fs)
	return sys
}

func (s *System) RunOneClock(rHook *readStateHook) {
	cpu := s.CPU
	if !cpu.Halted {
		cpu.Pipeline.RunOneClock()
		//time.Sleep(time.Millisecond * 100) // Sleep for 100 milliseconds to simulate clock cycles
	}
	if rHook != nil {
		(*rHook)(s)
	}
	if cpu.Halted {
		fmt.Println("CPU halted")
	}
}

func (s *System) RunToEnd(rHook *readStateHook) {
	for !s.CPU.Halted {
		s.RunOneClock(rHook)
	}
	s.CPU.PrintReg()
	s.CPU.RAM.PrintMem()
	s.CPU.Cache.PrintCache()
	fmt.Printf("PC: %d Cycles: %d\n", s.CPU.ProgramCounter,s.CPU.Clock)
	if rHook != nil {
		(*rHook)(s)
	}
}