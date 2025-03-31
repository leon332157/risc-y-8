package main

import (
	"fmt"

	"github.com/leon332157/risc-y-8/pkg/writeback"
	"github.com/leon332157/risc-y-8/pkg/memory"
	"github.com/leon332157/risc-y-8/pkg/alu"
	"github.com/leon332157/risc-y-8/pkg/decoder"
	"github.com/leon332157/risc-y-8/pkg/fetcher"
	"github.com/leon332157/risc-y-8/pkg/types"
)

func main() {

	// Create a RAM memory with 32 lines, 8 words per line, and 3 cycle delay
	ram_memory := memory.CreateRAM(32, 8, 3)
	
	// Create a cache with 8 sets, 2 ways, and no delay
	cache := memory.Default(&ram_memory)

	registers := alu.CreateRegisters()

	// Create an array of instructions to be in memory for demo
	inst_array := []uint32{


		// ADD r4, 16
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     4,
			ALU:    types.ImmALU["add"],
			Imm:    16,
		}).Encode(),

		// ADD r5, 32
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     5,
			ALU:    types.ImmALU["add"],
			Imm:    32,
		}).Encode(),

		// MUL r4, r5
		(&types.BaseInstruction{
			OpType: types.RegReg,
			Rd:     4,
			ALU:    types.RegALU["mul"],
			Rs:		5,
		}).Encode(),

		// STW r4, [r5 + 0x200]
		(&types.BaseInstruction{
			OpType: types.LoadStore,
			Rd:     4,
			Mode:   types.STW, 
			RMem:   5,
			Imm:	0x200,
		}).Encode(),

		// LDW r6, [r5 + 0x200]
		(&types.BaseInstruction{
			OpType: types.LoadStore,
			Rd:     6,
			Mode:   types.LDW,
			RMem:   5,
			Imm:    0x200,
		}).Encode(),

		// CMP r6, 0x200
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     6,
			ALU:    types.RegALU["cmp"],
			Imm:    0x200,
		}).Encode(),

		// BNE [pc + 4] skips the next sub instruction
		(&types.BaseInstruction{
			OpType: types.Control,
			RMem:   0,
			Flag:   types.Conditions["ne"].Flag,
			Mode:   types.Conditions["ne"].Mode,
			Imm:	4,
		}).Encode(),

		// SUB r4, 16
		(&types.BaseInstruction{
			OpType: types.RegImm,
			Rd:     4,
			ALU:    types.ImmALU["sub"],
			Imm:    16,
		}).Encode(),
	}

	// Copy the instructions from inst_array to the RAM memory
	copy(ram_memory.Contents, inst_array)

	for i := range inst_array {
		// Print the instruction in hex format
		fmt.Printf("Instruction %d: 0x%08x\n", i, ram_memory.Contents[i])
	}

	var wb_result types.WBToMem
	var mem_result types.MemToExe
	var alu_result types.ExeToDecode
	var decode_result types.DecodeToFetch
	var decode_input types.FetchToDecode
	var execute_input types.DecodeToExe
	var mem_input types.ExeToMem
	var wb_input types.MemToWB

	// pipeline loop until
	for {

		wb_result = writeback.WriteBackStage(registers, wb_input)
		fmt.Printf("\nWrite Back Stage to Memory Stage Result: %+v\n", wb_result)

		mem_result = memory.MemoryStageToExecute()
		fmt.Printf("\nMemory Stage to Execute Stage Result: %+v\n", mem_result)

		alu_result = alu.ExecuteStageToDecode()
		fmt.Printf("\nExecute Stage to Decode Stage Result: %+v\n", alu_result)

		decode_result = decoder.DecodeStageToFetch()
		fmt.Printf("\nDecode Stage to Fetch Stage Result: %+v\n", decode_result)

		decode_input = fetcher.FetchStageToDecode(ram_memory, registers)
		fmt.Printf("\nFetch Stage to Decode Stage Result: %+v\n", decode_input)

		execute_input = decoder.DecodeStageToExecute(decode_input)
		fmt.Printf("\nDecode Stage to Execute Stage Input: %+v\n", execute_input)

		mem_input = alu.ExecuteStageToMemory(registers, execute_input)
		fmt.Printf("\nExecute Stage to Memory Stage Input: %+v\n", mem_input)

		wb_input = memory.MemoryStageToWriteBack(mem_input, cache)
		fmt.Printf("\nMemory Stage to Write Back Stage Input: %+v\n", mem_input)

		if registers.IntRegisters[0] == 0 {
			fmt.Println("Program finished, exiting...")
			break
		}

	}

	// This works fine
	/* instruction := fetcher.FetchStageToDecode(ram_memory, registers)

	fmt.Printf("\nFetched Instruction: 0x%08x\n", instruction)
	fmt.Printf("Current PC: %d\n", registers.IntRegisters[0])

	decoded_inst := decoder.DecodeStageToExecute(instruction)

	fmt.Printf("Decoded Instruction: %+v\n", decoded_inst)

	alu.PrintIntegerRegisters(registers)

	mem_stage_input := alu.ExecuteStageToMemory(registers, decoded_inst)

	fmt.Printf("\nMemory Stage Input: %+v\n", mem_stage_input)

	wbsi := memory.MemoryStageToWriteBack(mem_stage_input, cache)

	fmt.Printf("\nWrite Back Stage Input: %+v\n", wbsi)

	writeback.WriteBackStage(registers, wbsi)

	alu.PrintIntegerRegisters(registers)
	alu.PrintRFlag(registers) */

}