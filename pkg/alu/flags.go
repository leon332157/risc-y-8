package alu

func (cpu *CPU) SetFlag(flag uint32) {
	cpu.RFlag |= flag
}

func (cpu *CPU) ClearFlag(flag uint32) {
	cpu.RFlag &^= flag
}

func (cpu *CPU) CheckFlag(flag uint32) bool {
	return (cpu.RFlag & flag) != 0
}

func (cpu *CPU) ResetFlags() {
	cpu.RFlag = 0
} 