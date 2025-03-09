package types

/* Registers (base 32 bit integer set):
16 32-bit General Purpose Registers 	(r1-r16) 0x01 - 0x10
	13 Integer Registers					(r1-r13)
	Base Pointer (bp)						(r14)
	Stack Pointer (sp)						(r15)
	Link Register (lr)						(r16)
Flag register 							(rflag) 0x11
Instruction Register 					(Separate from r1-r16)   (pc) can be referred by using 0x1F
Zero register 							(r0) 0x00 */

/* Registers (Float extension):
8 32-bit registers for Single Precision FP:
	8 Single precision fp 0x0 - 0x7	(f1-f8)
Floating point flags are not implemented and ignored. */

/* Registers (Vector extension):
8 128-bit Vec Register  0x0 - 0x7		(v1-v8)
Vector Flag Register (32 bits)			(vflag) (flags depend on type) can be referred by using 0x1D
Vector Type Register (32 bits) 0x1E		(vtype) (lower 16 bits are used for types) (upper bits are reserved)

Each 4 bits chunk corresponds to v0 to v8 respectively, from LSB to MSB.
Each 2 bits chunk corresponds to v0 to v8 respectively, from LSB to MSB.

VecType: 00, 4 packed ints, notation: i
VecType: 01, 4 packed float, notation: f
VecType: 10, 2 packed double. notation: d
VecType: 11, 16 packed bytes: notation: b */

var SpecialRegisters = map[uint8]string{
	0x00: "r0", 	0x11: "rflag", 	0x1D: "vflag", 	0x1E: "vtype",
	0x1F: "pc",
}

var IntegerRegisterNames = map[uint8]string{

	0x01: "r1", 	0x02: "r2", 	0x03: "r3", 	0x04: "r4",
	0x05: "r5", 	0x06: "r6", 	0x07: "r7", 	0x08: "r8",
	0x09: "r9", 	0x0A: "r10", 	0x0B: "r11", 	0x0C: "r12",
	0x0D: "r13", 	0x0E: "r14", 	0x0F: "r15", 	0x10: "r16",
}

var FPRegisterNames = map[uint8]string{
	0x00: "f1", 	0x01: "f2", 	0x02: "f3", 	0x03: "f4",
	0x04: "f5", 	0x05: "f6", 	0x06: "f7", 	0x07: "f8",
}

var VectorRegisterNames = map[uint8]string{
	0x00: "v1", 	0x01: "v2", 	0x02: "v3", 	0x03: "v4",
	0x04: "v5", 	0x05: "v6", 	0x06: "v7", 	0x07: "v8",
}

type CPU struct {

	// Special Registers
	ZeroRegister 	uint32
	RFlag 			uint32
	VFlag			uint32
	VType			uint32
	PC 				uint32

	// General Purpose Registers
	IntRegisters 	[16]uint32

	// Floating Point Registers
	FPRegisters 	[8]float32

	// Vector Registers
	VecRegisters	[8]uint32
}