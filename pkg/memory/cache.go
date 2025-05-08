package memory

import (
	"fmt"
	//"math"
	"math/bits"
)

type CacheType struct {
	Contents     [][]*CacheLine
	Sets         uint
	Ways         uint
	WordsPerLine uint
	LowerLevel   Memory
	MemoryRequestState
}

type CacheLine struct {
	Valid bool
	Tag   uint
	Data  []uint32
	LRU   int
}

type IdxTagOffs struct {
	index  uint
	tag    uint
	offset uint
}

func CreateCache(numSets, numWays, wordsPerLine, delay uint, lower Memory) CacheType {
	contents := make([][]*CacheLine, numSets)

	// initialize cache contents to zero
	for i := range numSets {
		contents[i] = make([]*CacheLine, numWays)
		lru := int(numWays - 1)
		for j := range numWays {
			data := make([]uint32, wordsPerLine)

			contents[i][j] = &CacheLine{Valid: false, Tag: 0, Data: data, LRU: lru}
			lru -= 1
		}
	}

	r := MemoryRequestState{
		NONE,  // No requester
		delay, // Delay cycles for servicing requests
		int(delay),
		false,
	}

	return CacheType{
		Contents:           contents,
		Sets:               numSets,
		Ways:               numWays,
		WordsPerLine:       wordsPerLine,
		LowerLevel:         lower,
		MemoryRequestState: r,
	}
}

// Creates the default cache
func CreateCacheDefault(lower Memory) CacheType {
	return CreateCache(8, 2, 4, 0, lower)
}

func (c *CacheType) IsBusy() bool {
	return c.MemoryRequestState.CyclesLeft > 0
}

func (c *CacheType) Requester() Requester {
	return c.MemoryRequestState.requester
}

func (c *CacheType) CancelRequest() {
	// Reset the request state
	c.MemoryRequestState = MemoryRequestState{
		NONE,
		c.MemoryRequestState.Delay,
		int(c.MemoryRequestState.Delay),
		false,
	}

	// c.sTracef("Cache cancelled request")
	// fmt.Printf("Cache MemoryRequestState is now %v \n", c.MemoryRequestState)
}

func (c *CacheType) service(who Requester) bool {
	if c.Sets == 0 || c.Ways == 0 {
		return true
	}
	if c.Delay == 0 {
		return true
	}
	// Check if memory is busy
	if c.MemoryRequestState.requester == NONE {
		// First request
		c.MemoryRequestState.requester = who
		c.MemoryRequestState.CyclesLeft = int(c.MemoryRequestState.Delay) // Reset the delay counter
		return false
	}
	if c.MemoryRequestState.CyclesLeft > 0 {
		if c.MemoryRequestState.requester == who {
			// if the same requester, then we decrement cycle left if cache is not waiting on memory
			if c.MemoryRequestState.WaitNext {
				return true
			} else {
				c.MemoryRequestState.CyclesLeft--
				if c.MemoryRequestState.CyclesLeft <= 0 {
					// c.MemoryRequestState.requester = NONE
					return true
				} else {
					return false
				}
			}
		} else {
			// different requester, cannot service
			return false
		}
	} else {
		return true
	}
	panic("oop cache")
	/* else {
		// Memory is idle, can service new request
		c.MemoryRequestState.CyclesLeft = int(c.MemoryRequestState.Delay) // Reset the delay counter
		c.MemoryRequestState.requester = who                              // Set the requester
		return true
	} */
}

func (c *CacheType) FindIndexTagOffset(addr uint) IdxTagOffs {
	// get lowest order 2 bit for data offset (reps 4 words)
	offsetBits := bits.Len32(uint32(c.WordsPerLine)) - 1
	offset := addr & ((1 << offsetBits) - 1)

	// Get set index from address
	indexBits := bits.Len32(uint32(c.Sets)) - 1
	index := (addr >> offsetBits) & ((1 << indexBits) - 1)

	// Get tag bits based on total bits needed for mem address
	//memSize := c.LowerLevel.SizeWords()
	totalBits := 32 //<- was originally this
	//totalBits := int(math.Log2(float64(memSize))) // Total bits needed for memory address
	tagBits := totalBits - indexBits - int(offsetBits)

	// Get tag using the address
	bin := uint(0b1)
	for range tagBits - 1 {
		bin = (bin << 1)
		bin += 1
	}
	tag := (addr >> uint(totalBits-tagBits)) & bin

	return IdxTagOffs{
		index:  index,
		tag:    tag,
		offset: offset,
	}
}

func (c *CacheType) Read(addr uint, who Requester) ReadResult {
	if who >= 0 {
		panic("Cache Read: Non-pipeline requester cannot read from cache")
	}

	if !c.service(who) {
		return ReadResult{WAIT, 0}
	}

	// If cache is disabled, read straight from memory
	if c.Sets == 0 || c.Ways == 0 || c.WordsPerLine == 0 {
		read := c.LowerLevel.Read(addr, who)
		return read
	}

	// Given the address, find the index of the set and tag
	ito := c.FindIndexTagOffset(addr)
	index, tag, offset := ito.index, ito.tag, ito.offset

	// If tag exists, check valid bit (false -> miss, true -> hit) cache hit: return the data, update lru
	set := c.Contents[index]
	for i := range c.Ways {
		if (set[i].Tag == tag) && (set[i].Valid) {
			c.UpdateLRU(index, i)
			return ReadResult{SUCCESS, set[i].Data[offset]}
		}
	}

	// Else, cache miss: read LINE from memory, load into cache, return data (no need to write back to mem)
	read := c.LowerLevel.ReadMulti(addr, c.WordsPerLine, offset, who)
	switch read.State {
	case WAIT:
		//fmt.Println("Cache miss, waiting for ram")
		c.MemoryRequestState.WaitNext = true
		return ReadResult{WAIT_NEXT_LEVEL, 0}
	case WAIT_NEXT_LEVEL:
		//fmt.Println("Cache miss, waiting for next level memory")
		c.MemoryRequestState.WaitNext = true
		return ReadResult{WAIT_NEXT_LEVEL, 0}
	case SUCCESS:
		c.CancelRequest() // reset the request state
	default:
		panic("AHHHH")
		//return ReadResult{read.State, 0}
		// do nothing
	}

	lruIdx := c.GetLRU(index)
	c.Contents[index][lruIdx] = &CacheLine{Valid: true, Tag: tag, Data: read.Value, LRU: c.Contents[index][lruIdx].LRU}
	c.UpdateLRU(index, lruIdx)

	return ReadResult{SUCCESS, read.Value[offset]}
}

// Write through, allocate policy
func (c *CacheType) Write(addr uint, who Requester, val uint32) WriteResult {
	if !c.service(who) {
		return WriteResult{WAIT, 0}
	}
	// If cache is disabled, write straight to memory
	if c.Sets == 0 || c.Ways == 0 {
		written := c.LowerLevel.Write(addr, who, val)
		return written
	}

	// Given address find the set index and tag
	ito := c.FindIndexTagOffset(addr)
	index, tag, offset := ito.index, ito.tag, ito.offset
	valid := false
	set := c.Contents[index]
	for i := range c.Ways {

		// If idx-tag exits, write to cache and write-through to memory
		curTag := set[i].Tag
		curValid := set[i].Valid

		if (curTag == tag) && curValid {

			// if data is the same, update lru, do nothing
			/*if set[i].Data == val {
				c.UpdateLRU(index, i)
				return SUCCESS
			}*/
			// if data is different, write-through to memory
			c.Contents[index][i].Data[offset] = val
			c.UpdateLRU(index, i)
			valid = true
		}
	}

	if !valid {
		// Find next empty line or LRU (empty line will be lru!), allocate in the cache
		lruIdx := c.GetLRU(index)
		d := c.Contents[index][lruIdx]
		d.Data[offset] = val
		c.Contents[index][lruIdx] = &CacheLine{Valid: true, Tag: tag, Data: d.Data, LRU: d.LRU} // keep lru the same, then use update function
		c.UpdateLRU(index, lruIdx)
	}

	// write through to memory
	written := c.LowerLevel.Write(addr, who, val)
	switch written.State {
	case WAIT, WAIT_NEXT_LEVEL:
		c.MemoryRequestState.WaitNext = true
		return WriteResult{WAIT_NEXT_LEVEL, 0} // Waiting for next level memory to service the request
	case SUCCESS:
		c.CancelRequest()
		return WriteResult{SUCCESS, written.Written} // Successfully wrote to memory (write-through)
	default:
		return WriteResult{FAILURE_INVALID_STATE, 0} // Failure to write to memory, return failure
	}
}

func (c *CacheType) UpdateLRU(setIndex uint, line uint) {
	set := c.Contents[setIndex]
	accessedLRU := set[line].LRU

	// Only update if not already MRU
	if accessedLRU != 0 {
		for i := range c.Ways {
			if i == line {
				set[i].LRU = 0 // Set accessed line to MRU
			} else if set[i].LRU < accessedLRU {
				set[i].LRU += 1 // Bump more-recently-used lines downward
			}
		}
	}
}


func (c *CacheType) GetLRU(setIndex uint) uint {
	set := c.Contents[setIndex]
	lru := -1
	lruIdx := -1

	for i := range c.Ways {
		if set[i].LRU > lru {
			lru = set[i].LRU
			lruIdx = int(i)
		}
	}
	if lru < 0 {
		panic("GetLRU: No valid LRU found in set index " + fmt.Sprint(setIndex))
	}
	return uint(lruIdx)
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
