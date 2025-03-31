package alu

func (cpu *Registers) SetFlag(flag uint32) {
	cpu.RFlag |= flag
}

func (cpu *Registers) ClearFlag(flag uint32) {
	cpu.RFlag &^= flag
}

func (cpu *Registers) CheckFlag(flag uint32) bool {
	return (cpu.RFlag & flag) != 0
}

func (cpu *Registers) ResetFlags() {
	cpu.RFlag = 0
} 