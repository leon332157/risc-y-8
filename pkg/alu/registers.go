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
	println("Integer Registers:")
	for i := 0; i < 32; i++ {
		fmt.Printf("r%d: %d\n", i+1, cpu.IntRegisters[i])
	}
}

func PrintFloatingPointRegisters(cpu *Registers) {
	println("Floating Point Registers:")
	for i := 0; i < 8; i++ {
		fmt.Printf("f%d: %.6f\n", i+1, cpu.FPRegisters[i])
	}
}

func PrintVectorRegisters(cpu *Registers) {
	println("Vector Registers:")
	for i := 0; i < 8; i++ {
		fmt.Printf("v%d: %d\n", i+1, cpu.VecRegisters[i])
	}
}

func PrintRFlag(cpu *Registers) {

	fmt.Printf("RFlag: %032b\n", cpu.RFlag)

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