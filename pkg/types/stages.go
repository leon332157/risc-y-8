package types

type MemoryStageInput struct {
	Address		int
	Data		uint32
	IsLoad		bool
	IsControl	bool
	IsALU		bool
	DestReg		uint8
	RegVal		uint32
	Flag		uint32
	PC			uint32	// if pc = 0, then no change to pc for branch
}