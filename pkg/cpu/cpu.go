package cpu

import (
	"fmt"
	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/memory"
	"github.com/leon332157/risc-y-8/pkg/types"
)

const (
	INT_REG_COUNT    = 32
	FLOAT_REG_COUNT  = 16
	VECTOR_REG_COUNT = 8
)

type IntRegister struct {
	Value       uint32 // Register value
	ReadEnable  bool   // Read enable flag
	WriteEnable bool   // Write enable flag
}

type FloatRegister struct {
	Value       float32 // Register value
	ReadEnable  bool    // Read enable flag
	WriteEnable bool    // Write enable flag
}

type VectorRegister struct {
	Value       [4]uint32 // Register value (4x 32-bit values)
	ReadEnable  bool      // Read enable flag
	WriteEnable bool      // Write enable flag
}

type CPU struct {
	Clock          uint32
	ProgramCounter uint32
	ALU            *alu.ALU
	Cache          *memory.CacheType
	RAM            *memory.RAM // Reference to RAM, if needed for direct access (optional)
	Pipeline       *Pipeline
	Flag           uint32 // RFlag
	IntRegisters   [INT_REG_COUNT]IntRegister
	//FloatRegisters  []FloatRegister
	//VectorRegisters []VectorRegister
	//FPU            *FPU
	//VPU            *VPU
}

func (cpu *CPU) blockRegister(r uint8) {
	if r >= uint8(len(cpu.IntRegisters)) {
		// Handle out of bounds access, if necessary
		panic("attempted to block an out of bounds register")
	}
	cpu.IntRegisters[r].ReadEnable = false
	cpu.IntRegisters[r].WriteEnable = false
}

func (cpu *CPU) unblockRegister(r uint8) {
	// Unblock the register for reading and writing
	if r >= uint8(len(cpu.IntRegisters)) {
		panic("attempted to unblock an out of bounds register") // Ensure we don't access out of bounds registers
	}
	cpu.IntRegisters[r].ReadEnable = true  // Allow reading from the register again
	cpu.IntRegisters[r].WriteEnable = true // Allow writing to the register again
}

func (cpu *CPU) Init(cache *memory.CacheType, ram *memory.RAM, p *Pipeline) {
	cpu.Clock = 0
	cpu.ProgramCounter = INIT_VECTOR
	cpu.ALU = alu.NewALU() // Create a new ALU instance
	cpu.Pipeline = p       // Set the pipeline reference
	cpu.Cache = cache
	cpu.RAM = ram
	for i := 0; i < INT_REG_COUNT; i++ {
		reg := &cpu.IntRegisters[i] // Get the pointer to the integer register
		reg.Value = 0               // Initialize all integer registers to 0
		reg.ReadEnable = true       // Allow reading by default
		reg.WriteEnable = true      // Allow writing by default
	}

}

func (cpu *CPU) PrintReg() {
	for i, reg := range cpu.IntRegisters {

		if i%8 == 0 && i != 0 {
			fmt.Println() // Newline for readability every 8 registers
		}
		fmt.Printf("r%d: 0x%08x\t", i, reg.Value)

	}
	fmt.Println()
}

const INIT_VECTOR = 0

func Main() {

	inst_array := []uint32{

		// ADD r4, 16
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     4,
			ALU:    types.ImmALU["add"],
			Imm:    16,
		}).Encode(),

		// ADD r5, 32
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     5,
			ALU:    types.ImmALU["add"],
			Imm:    32,
		}).Encode(),
		// ADD r5, 32
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     5,
			ALU:    types.ImmALU["add"],
			Imm:    32,
		}).Encode(),

		// ADD r5, 32
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     5,
			ALU:    types.ImmALU["add"],
			Imm:    32,
		}).Encode(),

		// ADD r5, 32
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     5,
			ALU:    types.ImmALU["add"],
			Imm:    32,
		}).Encode(),
		// ADD r5, 32
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     5,
			ALU:    types.ImmALU["add"],
			Imm:    32,
		}).Encode(),

		// STW r4, [r20 + 30]
		(&types.BaseInstruction{
			OpType: types.LoadStore,
			Rd:     4,
			Mode:   types.STW,
			RMem:   20,
			Imm:    30,
		}).Encode(),

		// LDW r6, [r20 + 29]
		(&types.BaseInstruction{
			OpType: types.LoadStore,
			Rd:     6,
			Mode:   types.LDW,
			RMem:   20,
			Imm:    29,
		}).Encode(),

		// STW r6, [r20 + 29]
		(&types.BaseInstruction{
			OpType: types.LoadStore,
			Rd:     6,
			Mode:   types.STW,
			RMem:   20,
			Imm:    29,
		}).Encode(),

		// CMP r6, 0
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     6,
			ALU:    types.RegALU["cmp"],
			Imm:    0,
		}).Encode(),

		// Beq [0] skips the next sub instruction
		(&types.BaseInstruction{
			OpType: types.Control,
			RMem:   20,
			Flag:   types.Conditions["eq"].Flag,
			Mode:   types.Conditions["eq"].Mode,
			Imm:    0,
		}).Encode(),

		// SUB r4, 16
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     4,
			ALU:    types.ImmALU["sub"],
			Imm:    1,
		}).Encode(),

		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     8,
			ALU:    types.ImmALU["add"],
			Imm:    2,
		}).Encode(),

		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     8,
			ALU:    types.ImmALU["add"],
			Imm:    2,
		}).Encode(),
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     8,
			ALU:    types.ImmALU["add"],
			Imm:    2,
		}).Encode(),
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     8,
			ALU:    types.ImmALU["add"],
			Imm:    2,
		}).Encode(),
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     8,
			ALU:    types.ImmALU["add"],
			Imm:    2,
		}).Encode(),
	}

	ram := memory.CreateRAM(32, 1, 4)

	copy(ram.Contents, inst_array)

	cache := memory.CreateCacheDefault(&ram) // Create a cache with default settings, 8 sets, 2 ways, no delay                  // Create a new ALU instance
	cpu := CPU{}
	pipeline := NewPipeline(&cpu)
	cpu.Init(&cache, &ram, pipeline) // Initialize the CPU with the cache and no pipeline yet
	fs := &FetchStage{}
	ds := &DecodeStage{}
	es := &ExecuteStage{}
	ms := &MemoryStage{}
	ws := &WriteBackStage{}
	fs.Init(pipeline, ds, nil)
	ds.Init(pipeline, es, fs)
	es.Init(pipeline, ms, ds)
	ms.Init(pipeline, ws, es)
	ws.Init(pipeline, nil, ms)
	pipeline.AddStages(ws, ms, es, ds, fs)
	for range 300 {
		pipeline.RunOnePass()
		fmt.Println("Clock:", cpu.Clock) // Print the clock cycle for debugging purposes
		fmt.Println("PC:", cpu.ProgramCounter)
		fmt.Println("Cache")
		cpu.Cache.PrintCache()
		fmt.Println("Memory")
		cpu.RAM.PrintMem()
		cpu.PrintReg()
	}
}
