package cpu

import (
	"fmt"
	"os"

	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/memory"
	"github.com/rs/zerolog"
)

const (
	INT_REG_COUNT    = 32
	FLOAT_REG_COUNT  = 16
	VECTOR_REG_COUNT = 8
)

const (
	READ_BLOCKED  = -1
	WRITE_BLOCKED = -2
)

type IntRegister struct {
	value       uint32 // Register value
	readEnable  bool   // Read enable flag
	writeEnable bool   // Write enable flag
}

type FloatRegister struct {
	value       float32 // Register value
	readEnable  bool    // Read enable flag
	writeEnable bool    // Write enable flag
}

type VectorRegister struct {
	Value       [4]uint32 // Register value (4x 32-bit values)
	ReadEnable  bool      // Read enable flag
	WriteEnable bool      // Write enable flag
}

type CPU struct {
	Clock          uint32
	ProgramCounter uint32
	Halted         bool
	ALU            *alu.ALU
	//FPU            *FPU
	//VPU            *VPU
	Cache        *memory.CacheType
	RAM          *memory.RAM // Reference to RAM, if needed for direct access (optional)
	Pipeline     *Pipeline
	IntRegisters [INT_REG_COUNT]IntRegister

	//FloatRegisters  []FloatRegister
	//VectorRegisters []VectorRegister
	log *zerolog.Logger
}

func (cpu *CPU) blockIntR(r uint8) {
	if r >= uint8(len(cpu.IntRegisters)) {
		// Handle out of bounds access, if necessary
		cpu.log.Panic().Msgf("attempted to block an out of bounds register: %v", r)
	}
	cpu.log.Trace().Msgf("Blocking register r%v for reading and writing", r)
	cpu.IntRegisters[r].readEnable = false
	cpu.IntRegisters[r].writeEnable = false
}

func (cpu *CPU) unblockIntR(r uint8) {
	// Unblock the register for reading and writing
	if r >= uint8(len(cpu.IntRegisters)) {
		cpu.log.Panic().Msgf("attempted to unblock an out of bounds register: %v", r)
	}
	cpu.log.Trace().Msgf("Unblocking register r%v for reading and writing", r)
	cpu.IntRegisters[r].readEnable = true
	cpu.IntRegisters[r].writeEnable = true
}

func (c *CPU) ReadIntR(r uint8) (uint32, int32) {
	if r == 0 {
		c.log.Info().Msg("attempted to read from r0, returning 0")
		return 0, SUCCESS // r0 is always 0
	}
	if r >= uint8(len(c.IntRegisters)) {
		c.log.Panic().Msgf("attempted to read an out of bounds register: %v", r)
	}
	if !c.IntRegisters[r].readEnable {
		return 0, READ_BLOCKED // Register is not readable
	}
	return c.IntRegisters[r].value, SUCCESS
}

func (c *CPU) ReadIntRNoBlock(r uint8) uint32 {
	if r == 0 {
		c.log.Info().Msg("attempted to read from r0, returning 0")
		return 0
	}
	if r >= uint8(len(c.IntRegisters)) {
		c.log.Panic().Msgf("attempted to read an out of bounds register: %v", r)
	}
	return c.IntRegisters[r].value
}

func (c *CPU) WriteIntRNoBlock(r uint8, v uint32) uint32 {
	if r == 0 {
		c.log.Info().Msg("attempted to write to r0, ignoring write")
		return 0 // r0 is always 0, ignore write
	}
	if r >= uint8(len(c.IntRegisters)) {
		c.log.Panic().Msgf("attempted to write an out of bounds register: %v", r)
	}
	c.IntRegisters[r].value = v
	return c.IntRegisters[r].value
}

func (c *CPU) WriteIntR(r uint8, value uint32) (uint32, int32) {
	if r == 0 {
		c.log.Info().Msg("attempted to write to r0, ignoring write")
		return 0, SUCCESS // r0 is always 0, ignore write
	}
	if r >= uint8(len(c.IntRegisters)) {
		c.log.Panic().Msgf("attempted to write an out of bounds register: %v", r)
	}
	if !c.IntRegisters[r].writeEnable {
		return 0, WRITE_BLOCKED // Register is not writable
	}
	c.IntRegisters[r].value = value
	return c.IntRegisters[r].value, SUCCESS
}

func (cpu *CPU) Init(cache *memory.CacheType, ram *memory.RAM, p *Pipeline, logger *zerolog.Logger) {
	cpu.Clock = 0
	cpu.ProgramCounter = INIT_VECTOR
	cpu.ALU = alu.NewALU() // Create a new ALU instance
	cpu.Pipeline = p       // Set the pipeline reference
	cpu.Cache = cache
	cpu.RAM = ram
	for i := 0; i < INT_REG_COUNT; i++ {
		reg := &cpu.IntRegisters[i] // Get the pointer to the integer register
		reg.value = 0               // Initialize all integer registers to 0
		reg.readEnable = true       // Allow reading by default
		reg.writeEnable = true      // Allow writing by default
	}
	l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	if logger == nil {
		logger = &l
	}
	cpu.log = logger
	cpu.log.Trace().Msgf("cpu initialized: %+v", &cpu)
}

func (cpu *CPU) PrintReg() {
	for i, reg := range cpu.IntRegisters {

		if i%8 == 0 && i != 0 {
			fmt.Println() // Newline for readability every 8 registers
		}
		fmt.Printf("r%d: 0x%08x\t", i, reg.value)

	}
	fmt.Println()
}

func (cpu *CPU) Halt() {
	cpu.Halted = true
	cpu.log.Info().Msg("CPU halted")
}

const INIT_VECTOR = 0

func CreateCPU() CPU {
	return CPU{}
}

/*
func Main() {

	inst_array := []uint32{

		// ADD r4, 16
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     4,
			ALU:    types.ImmALU["add"],
			Imm:    16,
		}).Encode(),

		(&types.BaseInstruction{
			OpType: types.Control,
			RMem:   0,
			Imm:    -1,
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
			OpType:  types.LoadStore,
			Rd:      4,
			MemMode: types.STW,
			RMem:    20,
			Imm:     30,
		}).Encode(),

		// LDW r6, [r20 + 29]
		(&types.BaseInstruction{
			OpType:  types.LoadStore,
			Rd:      6,
			MemMode: types.LDW,
			RMem:    20,
			Imm:     29,
		}).Encode(),

		// STW r6, [r20 + 29]
		(&types.BaseInstruction{
			OpType:  types.LoadStore,
			Rd:      6,
			MemMode: types.STW,
			RMem:    20,
			Imm:     29,
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
			OpType:   types.Control,
			RMem:     20,
			CtrlFlag: types.Conditions["eq"].Flag,
			MemMode:  types.Conditions["eq"].Mode,
			Imm:      0,
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
	ram := memory.CreateRAM(32, 1, 1)
	copy(make([]uint32, 0), inst_array)
	inststr := `
	ldi r5, 3
	ldi r6, 7
	ldi r2, 10
	add r4, 2
	sub r2, 1
	cmp r2, 0
	beq [r6]
	bne [r5]
	or r1, 0xbeef
	meow
	nop
	nop
	nop
	nop
	`
	instructions, err := r8.ParseString("cpuTest", inststr)
	if err != nil {
		panic(err)
	}
	copy(ram.Contents, instructions)

	cache := memory.CreateCacheDefault(&ram) // Create a cache with default settings, 8 sets, 2 ways, no delay                  // Create a new ALU instance
	cpu := CPU{}
	pipeline := NewPipeline(&cpu, false)
	clog := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	cpu.Init(&cache, &ram, pipeline, &clog) // Initialize the CPU with the cache and no pipeline yet
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
	for {
		if !cpu.halted {
			pipeline.RunOneClock()
		}
		fmt.Println("Clock:", cpu.Clock) // Print the clock cycle for debugging purposes
		fmt.Println("PC:", cpu.ProgramCounter)
		fmt.Println("Cache")
		cpu.Cache.PrintCache()
		fmt.Println("Memory")
		cpu.RAM.PrintMem()
		cpu.PrintReg()
		if cpu.halted {
			break
		}
	}
}
*/
