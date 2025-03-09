package alu

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