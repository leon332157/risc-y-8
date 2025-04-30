package memory

import (
	"fmt"
	"math"
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

type IdxTag struct {
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
		0,
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
	// get lowest order 2 bit for data offset (reps 4 words)
	offsetBits := bits.Len32(uint32(c.WordsPerLine)) - 1
	offset := addr & ((1 << offsetBits) - 1)

	// Get set index from address
	indexBits := bits.Len32(uint32(c.Sets)) - 1
	index := (addr >> offsetBits) & ((1 << indexBits) - 1)

	// Get tag bits based on total bits needed for mem address
	memSize := c.LowerLevel.SizeWords()
	// totalBits := 32 <- was originally this
	totalBits := int(math.Log2(float64(memSize))) // Total bits needed for memory address
	tagBits := totalBits - indexBits - int(offsetBits)

	// Get tag using the address
	bin := uint(0b1)
	for range tagBits - 1 {
		bin = (bin << 1)
		bin += 1
	}
	tag := (addr >> uint(totalBits-tagBits)) & bin

	return IdxTag{
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
	if c.Sets == 0 || c.Ways == 0 {
		read := c.LowerLevel.Read(addr, LAST_LEVEL_CACHE)
		switch read.State {
		case WAIT:
			//fmt.Println("Cache read, waiting for ram")
			return ReadResult{WAIT_NEXT_LEVEL, 0}
		case WAIT_NEXT_LEVEL:
			//fmt.Println("Cache read, waiting for next level memory")
			return ReadResult{WAIT_NEXT_LEVEL, 0}
		case SUCCESS:
			return read
		default:
			return ReadResult{read.State, 0}
			// do nothing
		}
	}

	// Given the address, find the index of the set and tag
	idxTag := c.FindIndexAndTag(addr)
	index, tag, offset := idxTag.index, idxTag.tag, idxTag.offset

	// If tag exists, check valid bit (false -> miss, true -> hit) cache hit: return the data, update lru
	set := c.Contents[index]
	for i := range c.Ways {
		if (set[i].Tag == tag) && (set[i].Valid) {
			c.UpdateLRU(index, i)
			return ReadResult{SUCCESS, set[i].Data[offset]}
		}
	}

	// Else, cache miss: read LINE from memory, load into cache, return data (no need to write back to mem)
	read := c.LowerLevel.ReadMulti(addr, c.WordsPerLine, offset, LAST_LEVEL_CACHE) // TODO: this is readMulti
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
	c.Contents[index][lruIdx] = &CacheLine{Valid: true, Tag: tag, Data: read.Value, LRU: c.Contents[index][lruIdx].LRU}
	c.UpdateLRU(index, lruIdx)

	return ReadResult{SUCCESS, read.Value[offset]}
}

// Write through, no allocate policy
func (c *CacheType) Write(addr uint, who Requester, val uint32) WriteResult {
	if !c.service(who) {
		return WriteResult{WAIT, 0}
	}
	// If cache is disabled, write straight to memory
	if c.Sets == 0 || c.Ways == 0 {
		written := c.LowerLevel.Write(addr, LAST_LEVEL_CACHE, val)
		switch written.State {
		case WAIT, WAIT_NEXT_LEVEL:
			return WriteResult{WAIT_NEXT_LEVEL, 0} // Waiting for next level memory to service the request
		case SUCCESS:
			return WriteResult{SUCCESS, written.Written} // Successfully wrote to memory (write-through)
		default:
			return WriteResult{FAILURE_INVALID_STATE, 0} // Failure to write to memory, return failure
		}
	}

	// Given address find the set index and tag
	idxTag := c.FindIndexAndTag(addr)
	index, tag, offset := idxTag.index, idxTag.tag, idxTag.offset
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
		// Find next empty line or LRU (empty line will be lru!), write-through to memory
		lruIdx := c.GetLRU(index)
		d := c.Contents[index][lruIdx]
		d.Data[offset] = val
		c.Contents[index][lruIdx] = &CacheLine{Valid: true, Tag: tag, Data: d.Data, LRU: d.LRU} // keep lru the same, then use update function
		c.UpdateLRU(index, lruIdx)
	}

	written := c.LowerLevel.Write(addr, LAST_LEVEL_CACHE, val)
	switch written.State {
	case WAIT, WAIT_NEXT_LEVEL:
		return WriteResult{WAIT_NEXT_LEVEL, 0} // Waiting for next level memory to service the request
	case SUCCESS:
		return WriteResult{SUCCESS, written.Written} // Successfully wrote to memory (write-through)
	default:
		return WriteResult{FAILURE_INVALID_STATE, 0} // Failure to write to memory, return failure
	}
}

func (c *CacheType) UpdateLRU(setIndex uint, line uint) {
	set := c.Contents[setIndex]

	// Only update lru if updated line's lru is not already zero
	if set[line].LRU != 0 {
		for i := range c.Ways {
			if i == line {
				c.Contents[setIndex][i].LRU = 0
			} else if set[i].LRU < int(c.Ways-1) {
				c.Contents[setIndex][i].LRU += 1
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
