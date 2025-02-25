package memory

// import "fmt" - checkout https://pkg.go.dev/fmt for more info
import "fmt"

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
	access MemoryAccess
	Contents [][]int
}

// RAM constructor, creates a new RAM instance
func createRAM(size, blockSize, wordSize, wordSize, latency int) *RAM {

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
		access: access,
		contents: contents,
	}
	
}