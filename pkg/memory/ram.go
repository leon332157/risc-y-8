package main

// import "fmt" - checkout https://pkg.go.dev/fmt for more info
import ("fmt")

// type to keep track of memory access
type AccessState struct {
	latency int
	accessed bool
}

// AccessControl constructor, creates a new AccessControl instance
func createAccessState(latency int) *AccessState {

	return &AccessState{
		latency: latency, 
		accessed: false,
	}
}

// Returns a bool to check if the mem has been accessed during the cycle
func (mem *AccessState) accessAttempt() bool {
	
	// If mem has been accessed, return false
	if mem.accessed {
		return false;
	}
	// If mem has not been accessed, access it and return true
	mem.accessed = true; 
	return true;
}

// Resets memory access state so it can be accessed again
func (mem *AccessState) resetAccessState() {
	mem.accessed = false;
}

// A line of memory or value in memory
type RAMValue struct {
	line []int
	value int
}

// RAM type with size and memory attributes
type RAM struct {
	size int
	blockSize int
	wordSize int
	access AccessState
	contents [][]int
}

// RAM constructor, creates a new RAM instance
func createRAM(size, blockSize, wordSize, latency int) *RAM {

	// Creates a slice (https://go.dev/tour/moretypes/7) to hold the RAM contents
	contents := make([][]int, size)

	// For each row of the RAM, make the necessary blocks
	for i := 0; i < size; i++ {
		contents[i] = make([]int, blockSize)
	}

	// Create new access state for the RAM
	access := createAccessState(latency)

	// returns address of the new RAM object
	return &RAM {
		size: size,
		blockSize: blockSize,
		wordSize: wordSize,
		access: &access,
		contents: contents,
	}	
}

// given an address, function that returns the aligned address (if needed)
func (mem *RAM) alignRAM(addr int) int {
	return ((addr % mem.size) / mem.wordSize) * mem.wordSize
}

// Calculates which memory block and offset within that block a given address corresponds to 
// returns the (block the word belongs to, where the word is inside the block)
func (mem *RAM) addrToOffset(addr int) (int, int) {
	alignedAddr := mem.alignRAM(addr)
	return (alignedAddr / mem.wordSize) % mem.size / mem.blockSize, (alignedAddr / mem.wordSize) % mem.blockSize
}

// MEMORY ACCESS READ AND WRITE

// Reads the value of the given address, can return entire line if lin is true
func (mem *RAM) read(addr int, lin bool) *RAMValue {

	// if memory cannot be accessed, it returns nothing (read more here about nil https://go101.org/article/nil.html)
	if !mem.access.accessAttempt() {
		return nil
	}

	// Reset access state for next cycle
	mem.access.resetAccessState()

	// gets block and offset addresses
	offset1, offset2 := mem.addrToOffset(addr)

	// If line is true, return the entire line
	if lin {
		return &RAMValue{line: append([]int{}, mem.contents[offset1]...)}
	}
	// else return the value at the address
	return &RAMValue{value: mem.contents[offset1][offset2]}
}

// Writes to RAM, returns a boolean depending on success of write
func (mem *RAM) write(addr int, val *RAMValue) bool {
	
	// memory cannot be accessed, return false
	if !mem.access.accessAttempt() {
		return false
	}

	// reset access state for next cycle
	mem.access.resetAccessState()

	// gets block and offset addresses
	offset1, offset2 := mem.addrToOffset(addr)

	// if val is a line, it writes to the entire line
	// else val is just a word, it writes a word
	if len(val.line) > 0 {
		mem.contents[offset1] = append([]int{}, val.line...)
	} else {
		mem.contents[offset1][offset2] = val.value
	}

	// successful write, return true
	return true
}

// Loads sequence of instructions into memory
func (mem *RAM) flash(instructions []int) {

	// For every instruction in the sequence
	for i := 0; i < len(instructions)*4; i += 4 {

		// get the block and offset address
		offset1, offset2 := mem.addrToOffset(i)
		// write to the address
		mem.contents[offset1][offset2] = instructions[i/4]
	}
}

// TEST
func main() {
	mem := createRAM(256, 4, 32, 5)
	mem.write(10, &RAMValue{value: 123})
	fmt.Println("Read: ", mem.read(10, false).value) // should print "Read: 123"
}