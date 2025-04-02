package cpu

import (
	"github.com/leon332157/risc-y-8/pkg/types"
)

type StageResult int;

const (
	SUCCESS = iota
	STALL = 1
	FAILURE = 2
	NOOP = 3
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
	Stages []Stage // List of pipeline stages
	cpu    *CPU  // Reference to the CPU instance
}

func (p *Pipeline) AddStage(stage Stage) {
	p.Stages = append(p.Stages, stage)
}

func (p *Pipeline) AddStages(stages... Stage) {
	for _, stage := range stages {
		if stage == nil {
			panic("[Pipeline AddStages] Attempted to add a nil stage")
		}
		p.AddStage(stage) // Add each stage individually
	}
}

func (p *Pipeline) Run() {

}

func (p *Pipeline) RunOnePass() {
	// wb -> mem -> exec -> dec -> fet
	for i:=0;i<len(p.Stages);i++ {
		p.Stages[i].Execute()
	}
	p.Stages[len(p.Stages)-1].Advance(nil, false) // Ensure the last stage can advance even if no instruction was passed to it, this is for the last stage in the pipeline (like WriteBack)
	p.cpu.Clock++
}

func (p *Pipeline) SquashALL() {
	for i := len(p.Stages) - 1; i >= 0; i-- {
		// Call Advance with a nil instruction and set stalled to true to squash the pipeline
		p.Stages[i].Advance(nil, true)
	}
}

type Stage interface {
	Init(pipeline *Pipeline, next Stage, prev Stage) error // Initialize the stage with a reference to the pipeline and the next stage
	Name() string // Get the name of the stage
	Execute() // Execute the stage logic
	Advance(i *InstructionIR, stalled bool)
}

type InstructionIR struct {
	rawInstruction uint32 // The instruction to be executed
	BaseInstruction types.BaseInstruction // Pointer to the base instruction structure, if applicable
	Op1 uint32
	Op2 uint32
	Result uint32
	DestMemAddr uint32 // ??
	RDestAux uint8 // Auxiliary register destination, used in some instructions (like PUSH,POP, CALL)
	ALUOp uint8
	MemOp uint8
	ControlFlag uint8
	ControlMode uint8
	WriteBack bool
}