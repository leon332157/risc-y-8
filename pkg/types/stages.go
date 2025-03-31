package types

type FetchToDecode struct {
	MemInst		uint32
}

type DecodeToExe struct {
	Instruction BaseInstruction
}

type ExeToMem struct {
	MemToWB
	Address		int
	Data		uint32
	IsLoad		bool
	IsControl	bool
	IsALU		bool
}

type MemToWB struct {
	Reg			uint8
	RegVal		uint32
	Branch_PC	uint32
	Flag		uint32
}

type WBToMem struct {
	Reg			uint8	// register that got written to
	RegVal		uint32	// value that was written to the register
}

type MemToExe struct {
	Address		int		// the address that was accessed
	Delay		int		// cycle delay
}

type ExeToDecode struct {
	Result		uint32
	Delay		int
}

type DecodeToFetch struct {
	Success		bool
}