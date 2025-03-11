package memory

import (
	"container/list" // for LRU queue
	"fmt"
)

// Create a memory interface
type Memory interface {
	createRAM(size, lineSize, wordSize, latency int) RAM
	read(addr int, lin bool) *RAMValue
	write(addr int, val *RAMValue) bool
	//flash(instructions []int)  // Might need later
}

// A line of memory or value in memory
type RAMValue struct {
	line  []uint32
	value uint32
}

// RAM type with size and memory attributes
type RAM struct {
	size     int         // number of lines in memory
	lineSize int         // number of words per line
	wordSize int         // number of bits per word
	access   AccessState // keep track of memory access
	contents [][]uint32  // data structure to hold memory contents
}

// type to keep track of memory access
type AccessState struct {
	latency    int
	cyclesLeft int
	accessed   bool
}

// AccessControl constructor, creates a new AccessControl instance
func createAccessState(latency int) *AccessState {

	return &AccessState{
		latency:    latency,
		cyclesLeft: latency,
		accessed:   false,
	}
}

// Returns a bool to check if the mem has been accessed during the cycle
func (c *AccessState) accessAttempt() bool {

	// If mem has been accessed, decrement cycles left and return false (must wait!)
	if c.accessed && c.cyclesLeft != 0 {
		c.cyclesLeft = c.cyclesLeft - 1
		return false
	}
	// If mem has not been accessed, access it, update the cycles left until next access and return true
	c.accessed = true
	c.cyclesLeft = c.latency
	return true
}

// MEMORY FUNCTIONS:

// RAM constructor, creates a new RAM instance
func createRAM(size, lineSize, wordSize, latency int) RAM {

	// Initialize contents: creates a slice of slice (https://go.dev/tour/moretypes/7) to hold the RAM contents
	contents := make([][]uint32, size)

	// For each row of the RAM, make the necessary cells
	for i := 0; i < size; i++ {
		contents[i] = make([]uint32, lineSize)
	}

	// Create new access state for the RAM
	access := createAccessState(latency)

	// returns address of the new RAM object
	return RAM{
		size:     size,
		lineSize: lineSize,
		wordSize: wordSize,
		access:   *access,
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
	return (alignedAddr / mem.wordSize) % mem.size / mem.lineSize, (alignedAddr / mem.wordSize) % mem.lineSize
}

// MEMORY ACCESS READ AND WRITE

// Reads the value of the given address, can return entire line if lin is true
func (mem *RAM) read(addr int, lin bool) *RAMValue {

	// if memory cannot be accessed, it returns nothing (read more here about nil https://go101.org/article/nil.html)
	if !mem.access.accessAttempt() {
		return nil
	}

	// Reset access state for next cycle
	//mem.access.resetAccessState()

	// gets block and offset addresses
	index, offset2 := mem.addrToOffset(addr)

	// If line is true, return the entire line
	if lin {
		return &RAMValue{line: append([]uint32{}, mem.contents[index]...)}
	}
	// else return the value at the address
	return &RAMValue{value: mem.contents[index][offset2]}
}

// Writes to RAM, returns a boolean depending on success of write
func (mem *RAM) write(addr int, val *RAMValue) bool {

	// memory cannot be accessed, return false
	if !mem.access.accessAttempt() {
		return false
	}

	// gets block and offset addresses
	offset1, offset2 := mem.addrToOffset(addr)

	// if val is a line, it writes to the entire line
	// else val is just a word, it writes a word
	if len(val.line) > 0 {
		mem.contents[offset1] = append([]uint32{}, val.line...)
	} else {
		mem.contents[offset1][offset2] = val.value
	}

	// successful write, return true
	return true
}

// Loads sequence of instructions into memory
func (mem *RAM) flash(instructions []uint32) {

	// For every instruction in the sequence
	for i := 0; i < len(instructions)*4; i += 4 {

		// get the block and offset address
		offset1, offset2 := mem.addrToOffset(i)
		// write to the address
		mem.contents[offset1][offset2] = instructions[i/4]
	}
}

// CACHE FUNCTIONS:

// Creates the default cache
func createDefault(mem RAM) CacheType {
	return configureCache(4, 4, 2, 0, mem)
}

// Creates a cache with configurable params
func configureCache(lineSize, numSets, ways, latency int, mem RAM) CacheType {

	access := createAccessState(latency)
	// initialize the sets
	sets := make([]Set, numSets)

	// Initialize empty sets with lru queue
	for i := range sets {
		sets[i] = Set{
			lines:    make([]CacheLine, ways),
			LRUQueue: list.New(),
		}
		// Initialize sets with empty lines
		for j := range sets[i].lines {
			sets[i].lines[j] = CacheLine{data: make([]uint32, lineSize), valid: false, dirty: false}
		}
	}

	// Return the new cache (return pointer instead??)
	return CacheType{
		lineSize: lineSize,
		numSets:  numSets,
		ways:     ways,
		access:   *access,
		sets:     sets,
		memory:   mem,
	}
}

// Checks cache access and depending on hit or miss, updates cache based on results
func (c *CacheType) search(addr int) bool {

	// If c can't be accessed, return falses
	if !c.access.accessAttempt() {
		return false
	}

	// Get the index, tag, from the address
	index := (addr / c.lineSize) % c.numSets
	tag := addr / (c.lineSize * c.numSets)
	set := &c.sets[index]

	// for each line in the set
	for i, line := range set.lines {
		// if its valid and the tag matches
		if line.valid && line.tag == tag {
			c.updateLRU(set, i)
			return true // Cache hit
		}
		if !line.valid {
			// TODO
		}
	}
	// Cache miss
	// TODO:
	// add data to first invalid line from memory
	//c.sets[index].lines[mn] =
	// OR if full, evict and replace (writeback)
	c.evictAndReplace(set, tag, addr)
	return true
}

// Update the LRU queue to see which line must be evicted next
func (c *CacheType) updateLRU(set *Set, idx int) {

	for e := set.LRUQueue.Front(); e != nil; e = e.Next() {
		if e.Value.(int) == idx {
			set.LRUQueue.Remove(e)
			break
		}
	}
	set.LRUQueue.PushFront(idx)
}

func (c *CacheType) evictAndReplace(set *Set, tag int, addr int) {

	victimIdx := c.getLRUVictim(set)
	set.lines[victimIdx].tag = tag
	set.lines[victimIdx].valid = true
	c.updateLRU(set, victimIdx)

	// TODO:
	// Read desired data from memory
	//toLoad := c.memory.read(addr, true)

	// delay?
	for range c.memory.access.latency {
		c.memory.read(addr, false)
	}
	// write evicted data to memory
	//c.memory.write(addr, )

	//delay

	// write new data in cache
}

func (c *CacheType) getLRUVictim(set *Set) int {

	elem := set.LRUQueue.Back()
	if elem != nil {
		return elem.Value.(int)
	}
	return 0
}

func main() {
	// TODO:
	mem := createRAM(256, 4, 32, 3)
	mem.write(10, &RAMValue{value: 123})
	for i := 0; i < 10; i += 1 {

		val := mem.read(10, false)

		if val == nil {
			fmt.Println("Read: ", "WAIT")
		} else {
			fmt.Println("Read: ", val.value) // should print "Read: 123"
		}
	}

	cache := createDefault(mem)
	print(cache)

	cache.search(10)
	print(cache)

	cache.search(32)
	print(cache)
}

func print(cache CacheType) {

	for i := 0; i < len(cache.sets); i++ {
		for j := 0; j < len(cache.sets[i].lines); j++ {
			fmt.Println(cache.sets[i].lines[j].data)
		}
	}
	fmt.Println("DONE")
}
