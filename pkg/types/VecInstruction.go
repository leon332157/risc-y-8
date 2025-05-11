package types

const (
	VPADD = iota
	VPSUB
	VPMUL
	VPSHL
	VPXOR
	VPAND
	VPORR
	VBEQ
)

const (
	VEC_LOAD_STORE uint8 = 0
	VEC_ARITH      uint8 = 1
)

var VPU map[string]uint8 = map[string]uint8{
	"vpadd": VPADD,
	"vpsub": VPSUB,
	"vpmul": VPMUL,
	"vpshl": VPSHL,
	"vpxor": VPXOR,
	"vpand": VPAND,
	"vpor":  VPORR,
	"vporr": VPORR,
	"vbeq":  VBEQ,
}

const (
	VEC_STORE_PACKED uint8 = iota
	VEC_LOAD_PACKED
)

var VecLoadStore map[string]uint8 = map[string]uint8{
	"vstp": VEC_STORE_PACKED,
	"vldp": VEC_LOAD_PACKED,
}

type VecInstruction struct {
	OpType  uint8 // 1 bit 0 load/store, 1 arth
	Vd      uint8 // Destination register
	Vs1     uint8 // Source register 1
	Vs2     uint8 // Source register 2
	VPU     uint8 // Vector Processing Unit
	MemMode uint8 // Load/store
	RMem    uint8
	Imm     int16 // 16 bit twos complement immediate value
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
