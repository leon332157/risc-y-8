package types

// Control
type ControlOp struct {
	Mode uint8 // 3 bits
	Flag uint8 // 4 bits
}

var (
	// ControlOp
	NE   = ControlOp{Mode: 0b000, Flag: 0b0000}
	NZ   = ControlOp{Mode: 0b000, Flag: 0b0000}
	EQ   = ControlOp{Mode: 0b000, Flag: 0b0001}
	Z    = ControlOp{Mode: 0b000, Flag: 0b0001}
	LT   = ControlOp{Mode: 0b001, Flag: 0b0110}
	GE   = ControlOp{Mode: 0b011, Flag: 0b0110}
	LU   = ControlOp{Mode: 0b100, Flag: 0b1000}
	AE   = ControlOp{Mode: 0b000, Flag: 0b1000}
	A    = ControlOp{Mode: 0b010, Flag: 0b1000}
	OF   = ControlOp{Mode: 0b100, Flag: 0b0100}
	NF   = ControlOp{Mode: 0b000, Flag: 0b0100}
	UNC  = ControlOp{Mode: 0b111, Flag: 0b0000}
	CALL = ControlOp{Mode: 0b111, Flag: 0b1111}
)

func GetModeFlag(c ControlOp) uint8 {
	return c.Mode<<4 | c.Flag
}

var Conditions = map[string]ControlOp{
	"ne":   NE,
	"nz":   NZ,
	"eq":   EQ,
	"z":    Z,
	"lt":   LT,
	"ge":   GE,
	"lu":   LU,
	"ae":   AE,
	"a":    A,
	"of":   OF,
	"nf":   NF,
	"unc":  UNC,
	"call": CALL,
}

const (
	IMM_ADD = iota
	IMM_SUB
	IMM_MUL
	IMM_AND
	IMM_XOR
	IMM_OR
	IMM_NOT
	IMM_NEG
	IMM_SHR
	IMM_SAR
	IMM_SHL
	IMM_ROL
	IMM_LDI
	IMM_LDX
	IMM_CMP
)

var ImmALU = map[string]uint8{
	"add": IMM_ADD,
	"sub": IMM_SUB,
	"mul": IMM_MUL,
	"and": IMM_AND,
	"xor": IMM_XOR,
	"orr": IMM_OR,
	"or":  IMM_OR,
	"not": IMM_NOT,
	"neg": IMM_NEG,
	"shr": IMM_SHR,
	"sar": IMM_SAR,
	"shl": IMM_SHL,
	"rol": IMM_ROL,
	"ror": IMM_ROL,
	"ldi": IMM_LDI,
	"ldx": IMM_LDX,
	"cmp": IMM_CMP,
}

var ImmALUInverse = map[uint8]string{}

const (
	REG_ADD = iota
	REG_SUB
	REG_MUL
	REG_DIV
	REG_REM
	REG_OR
	REG_XOR
	REG_AND
	REG_NOT
	REG_SHL
	REG_SHR
	REG_SAR
	REG_ROL
	REG_CMP
	REG_CPY
	REG_NSA
)

var RegALU = map[string]uint8{
	"add": REG_ADD,
	"sub": REG_SUB,
	"mul": REG_MUL,
	"div": REG_DIV,
	"rem": REG_REM,
	"orr": REG_OR,
	"or":  REG_OR,
	"xor": REG_XOR,
	"and": REG_AND,
	"not": REG_NOT,
	"shl": REG_SHL,
	"shr": REG_SHR,
	"sar": REG_SAR,
	"rol": REG_ROL,
	"cmp": REG_CMP,
	"cpy": REG_CPY,
	"mov": REG_CPY,
	"nsa": REG_NSA,
}

var RegALUInverse = map[uint8]string{}

func init() {
	for k, v := range ImmALU {
		ImmALUInverse[v] = k
	}
	for k, v := range RegALU {
		RegALUInverse[v] = k
	}
}

const (
	LDW  = iota // Load Word
	POP         // Pop
	PUSH        // Load X
	STW         // Store Word
)

type DataType uint8

const (
	//reserved	DataType = 0b00
	Integer DataType = 0b01
	Float   DataType = 0b10
	Vector  DataType = 0b11
)

const (
	RegImm    = 0b00 // reg-imm
	RegReg    = 0b01 // reg-reg
	LoadStore = 0b10 // load/store
	Control   = 0b11 // control
)

type BaseInstruction struct {
	OpType   uint8 // 00 for reg-imm, 01 for reg-reg, 10 for load/store, 11 for control
	Rd       uint8 // Destination register
	ALU      uint8 // ALU operation
	Rs       uint8 // Source register
	RMem     uint8 // Memory register
	CtrlFlag uint8 // Flag for control instructions
	CtrlMode uint8
	MemMode  uint8 // Mode for load/store instructions (0 for load, 10 pop, 10 for store, 11 push)
	Imm      int16 // 16 bit twos complement immediate value
}
