package memory

import (
	"fmt"
	"math"
)

type CacheType struct {
	Contents   [][]CacheLine
	Sets       uint
	Ways       uint
	LowerLevel Memory
	MemoryRequestState
}

type CacheLine struct {
	Valid bool
	Tag   uint
	Data  uint32
	LRU   int
}

type IdxTag struct {
	index uint
	tag   uint
}

func CreateCache(numSets, numWays, delay uint, lower Memory) CacheType {
	contents := make([][]CacheLine, numSets)

	// initialize cache contents to zero
	for i := range numSets {
		contents[i] = make([]CacheLine, numWays)
		lru := int(numWays - 1)
		for j := range numWays {
			contents[i][j] = CacheLine{Valid: false, Tag: 0, Data: 0, LRU: lru}
			lru -= 1
		}
	}

	r := MemoryRequestState{
		NONE,  // No requester
		delay, // Delay cycles for servicing requests
		0,
	}

	return CacheType{
		Contents:           contents,
		Sets:               numSets,
		Ways:               numWays,
		LowerLevel:         lower,
		MemoryRequestState: r,
	}
}

// Creates the default cache
func CreateCacheDefault(lower Memory) CacheType {
	return CreateCache(8, 4, 0, lower)
}

func (c *CacheType) IsBusy() bool {
	return c.MemoryRequestState.CyclesLeft > 0
}

func (c *CacheType) service(who Requester) bool {
	// Check if memory is busy
	if c.MemoryRequestState.CyclesLeft > 0 {
		if c.MemoryRequestState.requester == who {
			// if the same requester, then we decrement cycle left
			c.MemoryRequestState.CyclesLeft--
			return true // still servicing
		} else {
			// different requester, cannot service
			return false
		}
	} else {
		// Memory is idle, can service new request
		c.MemoryRequestState.CyclesLeft = int(c.MemoryRequestState.Delay) // Reset the delay counter
		c.MemoryRequestState.requester = who                              // Set the requester
		return true
	}
}

func (c *CacheType) FindIndexAndTag(addr uint) IdxTag {
	// Get set index from address
	index := addr % c.Sets
	// Get tag bits based on total bits needed for mem address
	memSize := c.LowerLevel.SizeWords()
	totalBits := int(math.Log2(float64(memSize)))
	tagBits := totalBits - int(math.Log2(float64(c.Sets)))

	// Get tag using the address
	bin := uint(0b1)
	for range tagBits - 1 {
		bin = (bin << 1)
		bin += 1
	}
	tag := (addr >> uint(totalBits-tagBits)) & bin

	return IdxTag{
		index: index,
		tag:   tag,
	}
}

func (c *CacheType) Read(addr uint, who Requester) ReadResult {
	if who >= 0 {
		panic("Cache Read: Non-pipeline requester cannot read from cache")
	}

	if !c.service(who) {
		return ReadResult{WAIT, 0}
	}

	// Given the address, find the index of the set and tag
	idxTag := c.FindIndexAndTag(addr)
	index, tag := idxTag.index, idxTag.tag

	// If tag exists, check valid bit (false -> miss, true -> hit) cache hit: return the data, update lru
	set := c.Contents[index]
	for i := range c.Ways {
		if (set[i].Tag == tag) && (set[i].Valid) {
			c.UpdateLRU(index, i)
			return ReadResult{SUCCESS, set[i].Data}
		}
	}

	// Else, cache miss: read memory, load into cache, return data (no need to write back to mem)
	read := c.LowerLevel.Read(addr, LAST_LEVEL_CACHE)
	switch read.State {
	case WAIT:
		//fmt.Println("Cache read, waiting for ram")
		return ReadResult{WAIT_NEXT_LEVEL, 0}
	case WAIT_NEXT_LEVEL:
		//fmt.Println("Cache read, waiting for next level memory")
		return ReadResult{WAIT_NEXT_LEVEL, 0}
	case SUCCESS:
	default:
		return ReadResult{read.State, 0}
		// do nothing
	}

	lruIdx := c.GetLRU(index)
	c.Contents[index][lruIdx] = CacheLine{Valid: true, Tag: tag, Data: read.Value, LRU: 0}
	c.UpdateLRU(index, lruIdx)

	return read
}

// Write through, no allocate policy
func (c *CacheType) Write(addr uint, who Requester, val uint32) WriteResult {
	if !c.service(who) {
		return WriteResult{WAIT, 0}
	}
	// Given address find the set index and tag
	idxTag := c.FindIndexAndTag(addr)
	index, tag := idxTag.index, idxTag.tag
	valid := false
	set := c.Contents[index]
	for i := range c.Ways - 1 {

		// If idx-tag exits, write to cache and write-through to memory
		if (set[i].Tag == tag) && set[i].Valid {

			// if data is the same, update lru, do nothing
			/*if set[i].Data == val {
				c.UpdateLRU(index, i)
				return SUCCESS
			}*/
			// if data is different, write-through to memory
			c.Contents[index][i].Data = val
			c.UpdateLRU(index, i)
			valid = true
		}
	}

	if !valid {
		// Find next empty line or LRU (empty line will be lru!), write-through to memory
		lruIdx := c.GetLRU(index)
		c.Contents[index][lruIdx] = CacheLine{Valid: true, Tag: tag, Data: val, LRU: 0}
		c.UpdateLRU(index, lruIdx)
	}

	written := c.LowerLevel.Write(addr, LAST_LEVEL_CACHE, val)
	switch written.State {
	case WAIT, WAIT_NEXT_LEVEL:
		return WriteResult{WAIT_NEXT_LEVEL, 0} // Waiting for next level memory to service the request
	case SUCCESS:
		return WriteResult{SUCCESS, written.Written} // Successfully wrote to memory (write-through)
	default:
		return WriteResult{FAILURE, 0} // Failure to write to memory, return failure
	}
	return WriteResult{FAILURE_INVALID_STATE, 0}
}

func (c *CacheType) UpdateLRU(setIndex uint, line uint) {
	set := c.Contents[setIndex]
	for i := range c.Ways {

		if i == line {
			c.Contents[setIndex][i].LRU = 0
		} else if set[i].LRU < int(c.Ways-1) {
			c.Contents[setIndex][i].LRU += 1
		}
	}
}

func (c *CacheType) GetLRU(setIndex uint) uint {
	set := c.Contents[setIndex]
	lru := -1

	for i := range c.Ways {
		if set[i].LRU > lru {
			lru = int(i)
		}
	}
	if lru < 0 {
		panic("GetLRU: No valid LRU found in set index " + fmt.Sprint(setIndex))
	}
	return uint(lru)
}

func (cache *CacheType) PrintCache() {
	fmt.Println("Tag    Index        Data    Valid    LRU")
	for i := range cache.Contents {
		for j := 0; j < len(cache.Contents[i]); j++ {
			line := cache.Contents[i][j]
			fmt.Printf("%05b    %03b    %08x    %t    %d\n", line.Tag, i, line.Data, line.Valid, line.LRU)
		} // might have to adjust depending on cache configs --> but nice looking for default cache
	}
	fmt.Println("")
}
