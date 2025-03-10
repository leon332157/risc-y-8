package cache

import (
	"container/list" // for LRU queue
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
}

type Set struct {
	lines    []CacheLine
	LRUQueue *list.List // Tracks LRU order with a queue
}

type CacheType struct {
	lineSize int
	numSets  int
	ways     int
	sets     []Set
}

// Creates the default cache
func createCache() CacheType {
	return configureCache(4, 4, 2)
}

// Creates a cache with configurable params
func configureCache(lineSize, numSets, ways int) CacheType {

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
			sets[i].lines[j] = CacheLine{data: make([]uint32, lineSize)}
		}
	}

	// Return the new cache (return pointer instead??)
	return CacheType{
		lineSize: lineSize,
		numSets:  numSets,
		ways:     ways,
		sets:     sets,
	}
}

// Checks if hit or miss, updates cache based on results
func (c *CacheType) search(addr int) bool {
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
	return false
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

	// Should I return the victim so that it can be written back to memory?
}

func (c *CacheType) getLRUVictim(set *Set) int {

	elem := set.LRUQueue.Back()
	if elem != nil {
		return elem.Value.(int)
	}
	return 0
}
