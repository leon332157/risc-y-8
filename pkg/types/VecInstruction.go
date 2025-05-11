package types

type VecInstruction struct {
	OpType uint8 // 1 bit 0 load/store, 1 arth
	Vd     uint8 // Destination register
	Vs1    uint8 // Source register 1
	Vs2    uint8 // Source register 2
	VPU    uint8 // Vector Processing Unit
	MemMode uint8 // Load/store
	RMem   uint8
	Scalar uint8  // Scalar bit
	Imm    uint16 // 16 bit twos complement immediate value
}

func LookUpVecOpType(opType uint8) string {
	switch opType {
	case 0:
		return "load/store"
	case 1:
		return "arth"
	default:
		return "unknown"
	}
}
