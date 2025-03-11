package cache

import (
	"container/list" // for LRU queue
	"fmt"
)

type Cache interface {
	createDefault()
	configureCache()
	search()
}

type CacheLine struct {
	tag   int
	data  []uint32
	valid bool
	dirty bool
}

type Set struct {
	lines    []CacheLine
	LRUQueue *list.List // Tracks LRU order with a queue
}

type CacheType struct {
	lineSize int
	numSets  int
	ways     int
	access   AccessState
	sets     []Set
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

// Resets cache access state so it can be accessed again
// func (c *AccessState) resetAccessState() {
// 	c.accessed = false
// 	c.cyclesLeft = c.latency
// }

// Creates the default cache
func createDefault() CacheType {
	return configureCache(4, 4, 2, 0)
}

// Creates a cache with configurable params
func configureCache(lineSize, numSets, ways, latency int) CacheType {

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
	}
	// Cache miss
	c.evictAndReplace(set, tag)
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

func (c *CacheType) evictAndReplace(set *Set, tag int) {

	victimIdx := c.getLRUVictim(set)
	set.lines[victimIdx].tag = tag
	set.lines[victimIdx].valid = true
	c.updateLRU(set, victimIdx)

	// must call writeback ---> cache and memory should be connected
	// Should I return the victim so that it can be written back to memory?
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
	cache := createDefault()
	print(cache)

	cache.search(10)
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
