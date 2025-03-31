package types

type MemoryStageInput struct {
	WriteBackStageInput
	Address		int
	Data		uint32
	IsLoad		bool
	IsControl	bool
	IsALU		bool
}

type WriteBackStageInput struct {
	Reg			uint8
	RegVal		uint32
	Branch_PC	uint32
	Flag		uint32
}