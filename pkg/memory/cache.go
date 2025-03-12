package memory

import (
	"container/list" // for LRU queue
	"fmt"
)

type Cache interface {
	createDefault(mem RAM) CacheType
	configureCache(lineSize, numSets, ways, latency int, mem RAM) CacheType
	search(addr int) bool
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
	memory   RAM
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

	return CacheType{
		lineSize: lineSize,
		numSets:  numSets,
		ways:     ways,
		access:   *access,
		sets:     sets,
		memory:   mem,
	}
}

// Checks cache access and depending on hit or miss, updates cache based on results (hit => true, miss => false)
func (c *CacheType) search(addr int) bool {

	// If c can't be accessed, return false
	if !c.access.AccessAttempt() {
		//fmt.Println("WAIT, cache can't be accessed")
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
	}

	// Cache miss
	// for each line in set, find the first invalid line
	for i, line := range set.lines {
		if !line.valid {
			// Add the new data to the invalid line
			set.lines[i] = CacheLine{
				tag:   tag,
				data:  c.memory.Read(addr, true).line,
				valid: true,
				dirty: false,
			}

			// update LRU queue
			c.updateLRU(set, i)
			return false
		}
	}

	// if no invalid line exists, must evict and replace
	// read replacement from mem
	replacement := c.memory.Read(addr, true).line

	// find least recently used line in the set and write victim back to memory
	victim := RAMValue{
		line: set.lines[c.getLRUVictim(set)].data,
	}
	c.memory.Write(addr, &victim)

	//store replacement to cache and flip the dirty bit
	set.lines[c.getLRUVictim(set)] = CacheLine{
		tag:   tag,
		data:  replacement,
		valid: true,
		dirty: true, // flip the dirty bit
	}

	return false
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
			fmt.Println(cache.sets[i].lines[j].data)
		}
	}
	fmt.Println("DONE")
}
