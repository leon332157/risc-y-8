package memory

import "fmt"

type RAM struct {
	Contents     []uint32
	NumLines     uint
	WordsPerLine uint
	MemoryRequestState
}

/* This function creates a new uint32 array of a certain size, lineSize, and delay.
*  	numLines int refers to the number of words the memory can hold (length of array)
   	wordsPerLine int refers to the number of words per line
   	delay int refers to the number of cycles that must be delayed between requests
*/
func CreateRAM(numLines uint, wordsPerLine uint, delay uint) RAM {

	// Create a slice
	size := numLines * wordsPerLine
	mem := make([]uint32, size)

	// Initialize memory to zero
	for i := range size {
		mem[i] = 0
	}

	r := MemoryRequestState{
		NONE,
		delay,
		int(delay),
	}

	return RAM{
		Contents:           mem,
		NumLines:           numLines,
		WordsPerLine:       wordsPerLine,
		MemoryRequestState: r,
	}
}

func (mem *RAM) IsBusy() bool {
	return mem.MemoryRequestState.CyclesLeft >= 0 
}

func (mem *RAM) service(who Requester) bool {
	if mem.MemoryRequestState.CyclesLeft > 0 { // if memory is busy rn
		if mem.MemoryRequestState.requester == who {
			// if the same requester, then we decrement cycle left
			mem.MemoryRequestState.CyclesLeft--
		} else {
			// different requester, cannot service
			if (mem.MemoryRequestState.requester == NONE)  {
				// If the memory is idle, we can service the new request
				mem.MemoryRequestState.CyclesLeft--
			}
			return false
		}
	} else {
		// Memory is idle, can service new request
		mem.MemoryRequestState.CyclesLeft = int(mem.MemoryRequestState.Delay) // Reset the delay counter
		mem.MemoryRequestState.requester = who                                // Set the requester
		return true
	}
	return false
}

// Reads a value from memory
func (mem *RAM) Read(addr uint, who Requester) ReadResult {
	if who <= 0 {
		// if not cache
		panic("Ram can not accept request from non-cache")
	}

	if !mem.service(who) { // if memory is busy, return WAIT
		return ReadResult{WAIT, 0} // Indicate that we are waiting
	}

	if addr > uint(len(mem.Contents)-1) {
		//fmt.Println("Address cannot be read. Not a valid address.")
		return ReadResult{FAILURE_OUT_OF_RANGE, 0}
	}

	return ReadResult{SUCCESS, mem.Contents[addr]}
}

// Writes a value to memory
func (mem *RAM) Write(addr uint, who Requester, val uint32) WriteResult {
	if who <= 0 {
		// if not cache
		panic("RAM can not accept request from non-cache")
	}

	if !mem.service(who) { // if memory is busy, return WAIT
		return WriteResult{WAIT, 0}// Indicate that we are waiting
	}

	if addr > uint(len(mem.Contents)-1) {
		//fmt.Println("Address cannot be read. Not a valid address.")
		return WriteResult{FAILURE_OUT_OF_RANGE,0}
	}

	mem.Contents[addr] = val
	return WriteResult{SUCCESS, 0}

}

func (mem *RAM) SizeBytes() uint {	
	return mem.NumLines * mem.WordsPerLine * 4 // 4 bytes per uint32
}

func (mem *RAM) SizeWords() uint {
	return mem.NumLines * mem.WordsPerLine
}

func (mem *RAM) SizeLines() uint {
	return mem.NumLines
}

func (mem *RAM) RequestState() MemoryRequestState {
	return mem.MemoryRequestState
}

// Prints memory
func (mem *RAM) PrintMem() {

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
			fmt.Print("] ")
		}

		if addr%8 == 0 && addr != 0 {
			fmt.Println("")
		}
	}
	fmt.Println()
}
