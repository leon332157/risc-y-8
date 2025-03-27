package memory

import (
	"container/list" // for LRU queue
	"fmt"
)

type Cache interface {
	Default(mem RAM) CacheType
	ConfigureCache(lineSize, numSets, ways, latency int, mem RAM) CacheType
	Search(addr int) bool
}

type CacheLine struct {
	tag   int
	data  []uint32
	valid bool
	dirty bool
}

type Set struct {
	lines    []CacheLine
	LRUQueue *list.List // Tracks LRU order for the sets lines with a queue
}

type CacheType struct {
	lineSize int
	numSets  int
	ways     int
	access   AccessState
	sets     []Set
	memory   *RAM
}

// CACHE FUNCTIONS:

// Creates the default cache
func CreateDefault(mem *RAM) CacheType {
	return ConfigureCache(4, 4, 2, 0, mem)
}

// Creates a cache with configurable params
func ConfigureCache(lineSize, numSets, ways, latency int, mem *RAM) CacheType {

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

	return CacheType{
		lineSize: lineSize,
		numSets:  numSets,
		ways:     ways,
		access:   *access,
		sets:     sets,
		memory:   mem,
	}
}

func (c *CacheType) Search(addr int) []uint32 {

	// i swear this same if writing/reading inProgress logic has been used like 3 or 4 times already just make it a function
	if c.memory.WriteInProgress {

		if c.memory.WriteCyclesLeft > 0 {
			c.memory.WriteCyclesLeft--
			fmt.Println("WAIT, memory write in progress. Cycles left:", c.memory.WriteCyclesLeft)
		}

		return nil

	} else if c.memory.ReadInProgress {

		if c.memory.Access.CyclesLeft > 0 {
			c.memory.Access.CyclesLeft--
			fmt.Println("WAIT, memory read in progress. Cycles left:", c.memory.Access.CyclesLeft)
		}

		return nil
	}

	index := (addr / c.lineSize) % c.numSets
	tag := addr / (c.lineSize * c.numSets)
	set := &c.sets[index]

	// Check if data is in cache
	for i, line := range set.lines {
		if line.valid && line.tag == tag {
			c.updateLRU(set, i)
			fmt.Println("Cache hit!")
			return line.data
		}
	}

	// Cache miss: Start memory read process
	fmt.Println("Cache miss. Requesting memory read at address:", addr)
	c.memory.StartRead(addr)

	return nil // No data yet, memory will complete later
}

func (c *CacheType) Insert(addr int, data []uint32) {
	index := (addr / c.lineSize) % c.numSets
	tag := addr / (c.lineSize * c.numSets)
	set := &c.sets[index]

	// Find an empty cache slot
	for i, line := range set.lines {
		if !line.valid {
			set.lines[i] = CacheLine{
				tag:   tag,
				data:  data,
				valid: true,
				dirty: false,
			}
			c.updateLRU(set, i)
			fmt.Println("Inserted into cache (empty slot).")
			return
		}
	}

	// No empty slot: Evict least recently used
	victimIdx := c.getLRUVictim(set)
	set.lines[victimIdx] = CacheLine{
		tag:   tag,
		data:  data,
		valid: true,
		dirty: false,
	}
	c.updateLRU(set, victimIdx)
	fmt.Println("Inserted into cache (evicted LRU line).")
}

// Update the LRU queue to see which line must be evicted next, pass most recently used as parameter
func (c *CacheType) updateLRU(set *Set, idx int) {

	for e := set.LRUQueue.Front(); e != nil; e = e.Next() {
		if e.Value.(int) == idx {
			set.LRUQueue.Remove(e)
			break
		}
	}
	set.LRUQueue.PushFront(idx)
}

func (c *CacheType) getLRUVictim(set *Set) int {

	elem := set.LRUQueue.Back()
	if elem != nil {
		return elem.Value.(int)
	}
	return 0
}

func PrintCache(cache CacheType) {

	for i := 0; i < len(cache.sets); i++ {
		for j := 0; j < len(cache.sets[i].lines); j++ {
			fmt.Printf("%08X\n", cache.sets[i].lines[j].data)
		}
	}
	fmt.Println("DONE")
	fmt.Println("")
}
