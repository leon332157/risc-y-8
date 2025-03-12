package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

import (
	"github.com/leon332157/risc-y-8/pkg/memory"
)

// print memory in hex
func Print2DSlice(slice [][]uint32) {
	for r, row := range slice {
		fmt.Printf("Row %d: ", r)
		for _, val := range row {
			fmt.Printf("0x%08X ", val)
		}
		fmt.Println()
	}
}

func main() {

	fmt.Println("Commands:\nstore <address> <value>\nload <address>\nread <address>\nview\nnext")
	fmt.Println("")

	mem := memory.CreateRAM(16, 4, 32, 5)

	reader := bufio.NewReader(os.Stdin)

	for {
		inp, _ := reader.ReadString('\n')
		inp = strings.TrimSpace(inp)
		stripped := strings.ToLower(inp)

		if stripped == "" || stripped == "next" {

			fmt.Println("Next cycle")
			continue

		}

		parts := strings.Split(stripped, " ")
		command := parts[0]

		switch command {
		case "store":

			if len(parts) != 3 {
				fmt.Println("Invalid command. Must be store <address> <value>")
				continue
			}

			// read hex value
			addr, err := strconv.ParseInt(parts[1], 0, 32)

			if err != nil {
				fmt.Println("Invalid address")
				continue
			}

			address := int(addr)

			val, err := strconv.ParseInt(parts[2], 0, 32)

			if err != nil {
				fmt.Println("Invalid value")
				continue
			}

			val_to_write := uint32(val)

			success := mem.Write(address, &memory.RAMValue{Line: []uint32{val_to_write, 0, 0, 0}})
			// success := mem.Write(4, &RAMValue{line: []uint32{0xDEADBEEF, 0xCAFEBABE, 0x12345678, 0x87654321}})

			if !success {
				fmt.Println("Failed to write to memory")
			}

			// Print2DSlice(mem.contents)

		case "load":

			if len(parts) != 2 {
				fmt.Println("Invalid command. Must be load <address>")
				continue
			}

			addrInt, err := strconv.ParseInt(parts[1], 0, 32)
			addr := int(addrInt)

			if err != nil {
				fmt.Println("Invalid address")
				continue
			}

			suc := mem.Read(addr, false)

			if suc == nil {
				continue
			}

			fmt.Printf("0x%08X\n", suc.Value)

		case "read":

			if len(parts) != 2 {
				fmt.Println("Invalid command. Must be read <address>")
				continue
			}

			addrInt, err := strconv.ParseInt(parts[1], 0, 32)

			if err != nil {
				fmt.Println("Invalid address")
				continue
			}

			fmt.Print("[")
			for i, v := range mem.Peek()[int(addrInt)] {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Printf("0x%08X", v)
			}
			fmt.Println("]")
		case "view":
			Print2DSlice(mem.Peek())
		default:
			fmt.Println("Unknown command. Must be either store, load, read, next")
		}
	}
}
