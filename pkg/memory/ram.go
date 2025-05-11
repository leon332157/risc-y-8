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
		false,
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

func (mem *RAM) Requester() Requester {
	return mem.MemoryRequestState.requester
}

func (mem *RAM) CancelRequest() {
	// Reset the request state
	mem.MemoryRequestState = MemoryRequestState{
		NONE, 
		mem.MemoryRequestState.Delay,	
		int(mem.MemoryRequestState.Delay),
		false,
	}

	// fmt.Println("RAM cancelled request")
}

func (mem *RAM) service(who Requester) bool {
	if mem.Delay == 0 {
		return true
	}
	if mem.MemoryRequestState.requester == NONE {
		// First request
		mem.requester = who
		mem.MemoryRequestState.CyclesLeft = int(mem.MemoryRequestState.Delay) // Reset the delay counter
		return false
	}
	if mem.MemoryRequestState.CyclesLeft > 0 { // if memory is busy rn
		if mem.MemoryRequestState.requester == who {
			// if the same requester, then we decrement cycle left
			mem.MemoryRequestState.CyclesLeft--
			if mem.MemoryRequestState.CyclesLeft == 0 {
				mem.MemoryRequestState.requester = NONE
				return true
			} else {
				return false
			}
		} else {
			// different requester, cannot service
			/*if mem.MemoryRequestState.requester == NONE {
				// If the memory is idle, we can service the new request
				mem.MemoryRequestState.CyclesLeft--
			}*/
			return false
		}
	}/*  else {
		// Memory is idle, can service new request
		mem.MemoryRequestState.CyclesLeft = int(mem.MemoryRequestState.Delay) // Reset the delay counter
		mem.MemoryRequestState.requester = who                                // Set the requester
		return true
	} */
	panic("oop ram")
}

// Reads a value from memory
func (mem *RAM) Read(addr uint, who Requester) ReadResult {
	
	if !mem.service(who) { // if memory is busy, return WAIT
		return ReadResult{WAIT, 0} // Indicate that we are waiting
	}

	if addr > uint(len(mem.Contents)-1) {
		//fmt.Println("Address cannot be read. Not a valid address.")
		return ReadResult{FAILURE_OUT_OF_RANGE, 0}
	}

	return ReadResult{SUCCESS, mem.Contents[addr]}
}

func (mem *RAM) ReadMulti(addr, numWords, offset uint, who Requester) ReadLineResult {
	
	if !mem.service(who) {
		return ReadLineResult{WAIT, []uint32{}}
	}

	if addr > uint(len(mem.Contents)-1) {
		fmt.Println("Address cannot be read. Not a valid address.")
		return ReadLineResult{FAILURE_OUT_OF_RANGE, []uint32{}}
	}

	a := addr - offset
	line := []uint32{}
	for i := range numWords {
		line = append(line, mem.Contents[a+i])
	}
	return ReadLineResult{SUCCESS, line}
}

// Writes a value to memory
func (mem *RAM) Write(addr uint, who Requester, val uint32) WriteResult {
	
	if !mem.service(who) { // if memory is busy, return WAIT
		return WriteResult{WAIT, 0} // Indicate that we are waiting
	}

	if addr > uint(len(mem.Contents)-1) {
		fmt.Println("Address cannot be read. Not a valid address.")
		return WriteResult{FAILURE_OUT_OF_RANGE, 0}
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
	for i := range uint(mem.NumLines) {
		row := []string{}
		header := fmt.Sprintf("0x%03X", i*mem.WordsPerLine)
		for range mem.WordsPerLine {
			row = append(row, fmt.Sprintf("0x%08X", mem.Contents[addr]))
			addr++
		}
		fmt.Printf("[%v] %v\n", header, row)
	}
}
