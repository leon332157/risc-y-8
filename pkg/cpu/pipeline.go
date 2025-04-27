package cpu

import (
	"fmt"
	"os"
	"strings"

	"github.com/leon332157/risc-y-8/pkg/types"
	"github.com/rs/zerolog"
)

type StageResult int

const (
	SUCCESS = iota
	STALL   = 1
	FAILURE = 2
	NOOP    = 3
)

func LookUpStageResult(s StageResult) string {
	switch s {
	case SUCCESS:
		return "SUCCESS"
	case STALL:
		return "STALL"
	case FAILURE:
		return "FAILURE"
	case NOOP:
		return "NOOP"
	default:
		return "UNKNOWN"
	}
}

type Pipeline struct {
	log        *zerolog.Logger // This logger is for stages
	pLog       *zerolog.Logger // This logger is for the pipeline itself
	Stages     []Stage         // List of pipeline stages
	cpu        *CPU            // Reference to the CPU instance
	canFetch   bool
	scalarMode bool // Flag to indicate if the pipeline is in scalar mode
}

func (p *Pipeline) AddStage(stage Stage) {
	p.Stages = append(p.Stages, stage)
}

func (p *Pipeline) AddStages(stages ...Stage) {
	for _, stage := range stages {
		if stage == nil {
			panic("[Pipeline AddStages] Attempted to add a nil stage")
		}
		p.AddStage(stage) // Add each stage individually
	}
}

func (p *Pipeline) Run() {

}

func (p *Pipeline) RunOneClock() {
	// wb -> mem -> exec -> dec -> fet
	p.RunBackPass()    // Run the backpass to execute the stages
	p.RunForwardPass() // Run the forward pass to advance the pipeline stages
}

func (p *Pipeline) RunBackPass() {
	for i := 0; i < len(p.Stages); i++ {
		p.Stages[i].Execute()
	}
	p.pLog.Trace().Msgf("canFetch: %v\n", p.canFetch)
}

func (p *Pipeline) RunForwardPass() {
	p.Stages[len(p.Stages)-1].Advance(nil, p.canFetch) // Ensure the last stage can advance even if no instruction was passed to it, this is for the last stage in the pipeline (like WriteBack)
	p.cpu.Clock++
}

func (p *Pipeline) WriteBackHook() {
	p.pLog.Info().Msg("WriteBackHook called")
	p.canFetch = true
}

func (p *Pipeline) SquashALL() {
	for i := len(p.Stages) - 1; i >= 0; i-- {
		// Call Advance with a nil instruction and set stalled to true to squash the pipeline
		p.Stages[i].Squash()
	}
}

func (p *Pipeline) sTrace(stage Stage, msg string) {
	if stage == nil {
		return
	}
	p.log.Trace().Str("Stage", (stage).Name()).Msg(msg)
}

func (p *Pipeline) sTracef(stage Stage, format string, args ...interface{}) {
	if stage == nil {
		return
	}
	p.log.Trace().Str("Stage", (stage).Name()).Msgf(format, args...)
}

func NewPipeline(cpu *CPU, scalar bool) *Pipeline {
	file, err := os.OpenFile("pipeline.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	logWriter := zerolog.ConsoleWriter{Out: file, NoColor: true} // Create a console writer for the log file
	log := zerolog.New(logWriter).Level(zerolog.TraceLevel).With().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + 1).Logger()
	pLog := zerolog.New(logWriter).With().Caller().Logger()
	p := Pipeline{
		Stages:     make([]Stage, 0),
		cpu:        cpu,
		log:        &log,
		pLog:       &pLog,
		canFetch:   true,
		scalarMode: scalar,
	}
	return &p
}

type Stage interface {
	Init(pipeline *Pipeline, next Stage, prev Stage) error // Initialize the stage with a reference to the pipeline and the next stage
	Name() string                                          // Get the name of the stage
	Execute()                                              // Execute the stage logic
	Advance(i *InstructionIR, stalled bool) bool           // Advance the stage with the current instruction and stalled status, return true if the stage advanced, false if it was stalled
	Squash() bool                                          // Squash the instruction in the stage, return true if the stage was squashed
	CanAdvance() bool                                      // Check if the stage can take in new instruction
	FormatInstruction() string                             // for TUI
}

type AssembledInst interface {
}
type InstructionIR struct {
	BaseInstruction *types.BaseInstruction // base instruction structure, if applicable
	//FloatInstruction types.FloatInstruction
	//VecInstruction types.VecInstruction
	Operand        uint32
	Result         uint32 // Calculated as "result = Rd op Rs/imm"
	RDestAux       uint8  // Auxiliary register destination, used in some instructions (like PUSH, POP, CALL)
	DestMemAddr    uint32 // Memory address for load/store operations, and branch destination
	BranchTaken    bool
	WriteBack      bool
	rawInstruction uint32 // The instruction to be executed
}

func (i *InstructionIR) FormatLines() string {
	if i == nil {
		return "<bubble>"
	}
	s := fmt.Sprintf("raw: %x\n", i.rawInstruction)
	if i.BaseInstruction == nil {
		s += "BaseInstruction: <nil>\n"
		return s
	}
	s += fmt.Sprintf("OpType: %x\n", i.BaseInstruction.OpType)
	switch i.BaseInstruction.OpType {
	case types.RegReg, types.RegImm:
		s += fmt.Sprintf("Rd: %x\n", i.BaseInstruction.Rd)
		s += fmt.Sprintf("ALU: %x\n", i.BaseInstruction.ALU)
		s += fmt.Sprintf("Rs: %x\n", i.BaseInstruction.Rs)
	case types.Control:
		s += fmt.Sprintf("RMem: %x\n", i.BaseInstruction.RMem)
		s += fmt.Sprintf("DestMemAddr: %x\n", i.DestMemAddr)
		s += fmt.Sprintf("RDestAux: %x\n", i.RDestAux)
		s += fmt.Sprintf("BranchTaken: %v\n", i.BranchTaken)
		s += fmt.Sprintf("BranchModeFlag: %x\n", i.BaseInstruction.CtrlMode<<4|i.BaseInstruction.CtrlFlag)
	case types.LoadStore:
		s += fmt.Sprintf("Rd: %x\n", i.BaseInstruction.Rd)
		s += fmt.Sprintf("RMem: %x\n", i.BaseInstruction.RMem)
		s += fmt.Sprintf("DestMemAddr: %x\n", i.DestMemAddr)
		s += fmt.Sprintf("MemMode: %x\n", i.BaseInstruction.MemMode)
	default:
	}

	return strings.Join([]string{s}, "\n")
}
