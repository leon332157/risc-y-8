package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/leon332157/risc-y-8/pkg/memory"
)

func main() {
	fmt.Println("Commands:\nstore <address> <value>\nload <address>\nread <address>\nnext\nview\ncache")
	fmt.Println("")

	mem := memory.CreateRAM(32, 8, 5)
	cache := memory.Default(&mem)

	reader := bufio.NewReader(os.Stdin)

	for {
		inp, _ := reader.ReadString('\n')
		inp = strings.TrimSpace(inp)
		stripped := strings.ToLower(inp)

		if stripped == "" || stripped == "next" {
			fmt.Println("Next cycle")

			// Handle memory write completion
			// if mem.WriteInProgress {
			// 	if mem.WriteCyclesLeft > 0 {
			// 		mem.WriteCyclesLeft--
			// 		fmt.Println("WAIT, memory write in progress. Cycles left:", mem.WriteCyclesLeft)
			// 		continue
			// 	}
			// 	mem.Contents[mem.WriteAddr] = mem.WriteData
			// 	mem.WriteInProgress = false
			// 	fmt.Printf("\nMemory write completed, wrote %08X to address %d\n", mem.WriteData, mem.WriteAddr)
			// 	fmt.Println("")
			// 	continue
			// }

			// Handle memory read completion
			// if mem.ReadInProgress {
			// 	if mem.Access.CyclesLeft > 0 {
			// 		mem.Access.CyclesLeft--
			// 		fmt.Println("WAIT, memory read in progress. Cycles left:", mem.Access.CyclesLeft)
			// 		continue
			// 	}

			// 	// Read from memory and insert into cache
			// 	fetchedValue := mem.Read(mem.LastReadAddr, true)
			// 	if fetchedValue != nil {
			// 		fmt.Printf("\nMemory read completed. Data: %08X\n", fetchedValue.Line)
			// 		cache.Insert(mem.LastReadAddr, fetchedValue.Line) // New function to insert into cache
			// 		fmt.Println("Memory read completed. Data loaded into cache.")
			// 	}
			// 	mem.ReadInProgress = false
			// 	continue
			// }

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

			addr, err := strconv.ParseInt(parts[1], 0, 32)
			if err != nil {
				fmt.Println("Invalid address")
				continue
			}

			val, err := strconv.ParseUint(parts[2], 0, 32)
			if err != nil {
				fmt.Println("Invalid value")
				continue
			}

			// since its write-through, no-allocate, we can write directly to memory
			mem.Write(int(addr), uint32(val))

		case "load":
			if len(parts) != 2 {
				fmt.Println("Invalid command. Must be load <address>")
				continue
			}

			addr, err := strconv.ParseInt(parts[1], 0, 32)
			if err != nil {
				fmt.Println("Invalid address")
				continue
			}

			res := cache.Read(int(addr))

			fmt.Printf("Value loaded from Cache: 0x%08X\n", res)

			// res := mem.Read(int(addr), false)
			// if res == nil {
			//  continue
			// }

			// fmt.Printf("0x%08X\n", res.value)

		case "read":
			if len(parts) != 2 {
				fmt.Println("Invalid command. Must be read <address>")
				continue
			}

			addr, err := strconv.ParseInt(parts[1], 0, 32)
			if err != nil {
				fmt.Println("Invalid address")
				continue
			}
			fmt.Print("Valid Addr", addr)

			fmt.Print("[")
			for i, v := range mem.Contents {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Printf("0x%08X", v)
			}
			fmt.Println("]")
			fmt.Println("")

		case "view":
			mem.PrintMem()
			fmt.Println("")

		case "cache":
			cache.PrintCache()

		default:
			fmt.Println("Unknown command. Must be either store, load, read, next, view, cache")
		}
	}
}
