package main

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"strconv"

	"pkg/memory/main.go"
)

func main() {

	var inp string

	reader := bufio.NewReader(os.Stdin)
	inp, _ = reader.ReadString('\n')
	inp = strings.TrimSpace(inp)
	stripped := strings.ToLower(inp)

	for  /* the instructions are still being run */ {

		if strings.HasPrefix(stripped, "peak") {

			parts := strings.Split(stripped, " ")
			load_address_str := parts[1]
			load_addr, err := strconv.ParseUint(load_address_str, 0, 32)

			if (err != nil) {
				fmt.Println("error")
				return
			}

			value := ram.read(int(load_addr), false)

			if value != nil {
				fmt.Println("Value at address", load_addr, ":", value.value)
			} else fmt.Println("Unable to access memory at", load_addr)

		} else stripped == "next" || stripped == "" {
			fmt.Println("continue 1 cycle")
			// continue by 1 cycle
		} 

		inp, _ = reader.ReadString('\n')
		inp = strings.TrimSpace(inp)
		stripped = strings.ToLower(inp)

	}

}