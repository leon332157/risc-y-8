package alu

import (
	"fmt"
)

type Registers struct {

	// Special Registers
	RFlag 			uint32

	// General Purpose Registers
	IntRegisters 	[32]uint32

	// Floating Point Registers
	FPRegisters 	[8]float32

	// Vector Registers
	VecRegisters	[8]uint32


}

func CreateRegisters() *Registers {
	// Create a new register set
	cpu := &Registers{}

	// Initialize all registers to 0
	for i := 0; i < 32; i++ {
		cpu.IntRegisters[i] = 0
	}
	for i := 0; i < 8; i++ {
		cpu.FPRegisters[i] = 0.0
		cpu.VecRegisters[i] = 0
	}
	
	// Initialize RFlag to 0
	cpu.RFlag = 0

	return cpu
}

func PrintIntegerRegisters(cpu *Registers) {
	fmt.Println("\nInteger Registers:")

	for i := range 4 {
		fmt.Printf("r%d:\t0x%08x\tr%d:\t0x%08x\tr%d:\t0x%08x\tr%d:\t0x%08x\tr%d:\t0x%08x\tr%d:\t0x%08x\tr%d:\t0x%08x\tr%d:\t0x%08x\n",
			i*8+0, cpu.IntRegisters[i*8+0],
			i*8+1, cpu.IntRegisters[i*8+1],
			i*8+2, cpu.IntRegisters[i*8+2],
			i*8+3, cpu.IntRegisters[i*8+3],
			i*8+4, cpu.IntRegisters[i*8+4],
			i*8+5, cpu.IntRegisters[i*8+5],
			i*8+6, cpu.IntRegisters[i*8+6],
			i*8+7, cpu.IntRegisters[i*8+7])
	}
}

func PrintFloatingPointRegisters(cpu *Registers) {
	fmt.Println("\nFloating Point Registers:")

	fmt.Printf("fp1:\t0x%08x\tfp2:\t0x%08x\tfp3:\t0x%08x\tfp4:\t0x%08x\tfp5:\t0x%08x\tfp6:\t0x%08x\tfp7:\t0x%08x\tfp8:\t0x%08x\n",
		uint32(cpu.FPRegisters[0]),
		uint32(cpu.FPRegisters[1]),
		uint32(cpu.FPRegisters[2]),
		uint32(cpu.FPRegisters[3]),
		uint32(cpu.FPRegisters[4]),
		uint32(cpu.FPRegisters[5]),
		uint32(cpu.FPRegisters[6]),
		uint32(cpu.FPRegisters[7]),
	)
}

func PrintVectorRegisters(cpu *Registers) {
	println("\nVector Registers:")
	fmt.Printf("vec1:\t0x%08x\tvec2:\t0x%08x\tvec3:\t0x%08x\tvec4:\t0x%08x\tvec5:\t0x%08x\tvec6:\t0x%08x\tvec7:\t0x%08x\tvec8:\t0x%08x\n",
		cpu.VecRegisters[0],
		cpu.VecRegisters[1],
		cpu.VecRegisters[2],
		cpu.VecRegisters[3],
		cpu.VecRegisters[4],
		cpu.VecRegisters[5],
		cpu.VecRegisters[6],
		cpu.VecRegisters[7],
	)
}

func PrintRFlag(cpu *Registers) {

	fmt.Printf("\nRFlag: %032b\n", cpu.RFlag)

	fmt.Println("Flags:")
	
	if cpu.CheckFlag(0x1) {
		fmt.Println("CF: Carry Flag")
	}

	if cpu.CheckFlag(0x2) {
		fmt.Println("ZF: Zero Flag")
	}

	if cpu.CheckFlag(0x4) {
		fmt.Println("SF: Sign Flag")
	}

	if cpu.CheckFlag(0x8) {
		fmt.Println("OVF: Overflow Flag")
	}
}

func PrintAllRegisters(cpu *Registers) {
	PrintIntegerRegisters(cpu)
	PrintFloatingPointRegisters(cpu)
	PrintVectorRegisters(cpu)
	PrintRFlag(cpu)
}