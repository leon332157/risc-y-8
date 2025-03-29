package memory

import "fmt"

// Create a memory interface
type Memory interface {
	CreateRAM(numLines int, wordsPerLine int, delay int) []uint32
	Read(addr int) uint32
	Write(addr int, val uint32) bool
}

type RAM struct {
	Contents     []uint32
	NumLines     int
	WordsPerLine int
	Delay        int
	Busy         bool
}

/* This function creates a new uint32 array of a certain size, lineSize, and delay.
*  	numLines int refers to the number of words the memory can hold (length of array)
   	wordsPerLine int refers to the number of words per line
   	delay int refers to the number of cycles that must be delayed between requests
*/
func CreateRAM(numLines int, wordsPerLine int, delay int) RAM {

	// Create a slice
	size := numLines * wordsPerLine
	mem := make([]uint32, size)

	// Initialize memory to zero
	for i := range size {
		mem[i] = 0
	}

	return RAM{
		Contents:     mem,
		NumLines:     numLines,
		WordsPerLine: wordsPerLine,
		Delay:        delay,
		Busy:         false, // if a stage has requested, flip to true
	}
}

// Reads a value from memory
func (mem RAM) Read(addr int) uint32 {

	// Make sure address is valid
	if addr > (len(mem.Contents) - 1) {
		fmt.Println("Address cannot be read. Not a valid address.")
		panic(mem.Contents[addr])
	}
	return mem.Contents[addr]
}

// Writes a value to memory
func (mem RAM) Write(addr int, val uint32) bool {

	// Make sure address is valid
	if addr > (len(mem.Contents) - 1) {
		fmt.Println("Address cannot be written to. Not a valid address.")
		panic(mem.Contents[addr])
	}

	mem.Contents[addr] = val
	return true
}

// Prints memory
func (mem RAM) PrintMem() {

	addr := 0
	for i := range mem.NumLines {
		if i < 10 {
			fmt.Print(i, "  [ ")
		} else {
			fmt.Print(i, " [ ")
		}

		for range mem.WordsPerLine {
			fmt.Printf("%08x ", mem.Contents[addr])
			addr++
		}

		if addr%8 == 0 && addr != 0 {
			fmt.Println("]")
		}
	}
}
