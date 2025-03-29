package memory

import (
	"fmt"
	"math"
)

type Cache interface {
	Default(mem RAM) CacheType
	CreateCache(numSets, numWays, delay int, mem RAM) CacheType
	Read(addr int) uint32
	Write(addr int, val uint) bool
}

type CacheType struct {
	Contents [][]CacheLine
	Sets     int
	Ways     int
	Delay    int
	Memory   RAM
}

type CacheLine struct {
	Valid bool
	Tag   int
	Data  uint32
	LRU   int
}

type IdxTag struct {
	index int
	tag   int
}

func CreateCache(numSets, numWays, delay int, mem RAM) CacheType {
	contents := make([][]CacheLine, numSets)

	// initialize cache contents to zero
	for i := range numSets {
		contents[i] = make([]CacheLine, numWays)
		lru := numWays - 1
		for j := range numWays {
			contents[i][j] = CacheLine{Valid: false, Tag: 0, Data: 0, LRU: lru}
			lru -= 1
		}
	}

	return CacheType{
		Contents: contents,
		Sets:     numSets,
		Ways:     numWays,
		Delay:    delay,
		Memory:   mem,
	}
}

// Creates the default cache
func Default(mem *RAM) CacheType {
	return CreateCache(8, 2, 0, *mem)
}

func (c *CacheType) FindIndexAndTag(addr int) IdxTag {
	// Get set index from address
	index := addr % c.Sets

	// Get tag bits based on total bits needed for mem address
	memSize := c.Memory.NumLines * c.Memory.WordsPerLine
	totalBits := int(math.Log2(float64(memSize)))
	tagBits := totalBits - int(math.Log2(float64(c.Sets)))

	// Get tag using the address
	bin := 0b1
	for range tagBits - 1 {
		bin = (bin << 1)
		bin += 1
	}
	tag := (addr >> (totalBits - tagBits)) & bin

	return IdxTag{
		index: index,
		tag:   tag,
	}
}

func (c *CacheType) Read(addr int) uint32 {
	// Given the address, find the index of the set and tag
	idxTag := c.FindIndexAndTag(addr)
	index, tag := idxTag.index, idxTag.tag

	// If tag exists, check valid bit (false -> miss, true -> hit) cache hit: return the data, update lru
	set := c.Contents[index]
	for i := range c.Ways {

		if set[i].Tag == tag && set[i].Valid {
			c.UpdateLRU(index, i)
			return set[i].Data
		}
	}
	// Else, cache miss: read memory, load into cache, return data (no need to write back to mem)
	replacement := c.Memory.Read(addr)
	lruIdx := c.GetLRU(index)
	c.Contents[index][lruIdx] = CacheLine{Valid: true, Tag: tag, Data: replacement, LRU: 0}
	c.UpdateLRU(index, lruIdx)

	return replacement
}

// Write through, no allocate policy
func (c *CacheType) Write(addr int, val uint32) bool {
	// Given address find the set index and tag
	idxTag := c.FindIndexAndTag(addr)
	index, tag := idxTag.index, idxTag.tag

	set := c.Contents[index]
	for i := range c.Ways - 1 {

		// If idx-tag exits, write to cache and write-through to memory
		if set[i].Tag == tag && set[i].Valid {

			// if data is the same, update lru, do nothing
			if set[i].Data == val {
				c.UpdateLRU(index, i)
				return true
			}
			// if data is different, write-through to memory
			c.Contents[index][i].Data = val
			c.UpdateLRU(index, i)
			c.Memory.Write(addr, val)
			return true
		}
	}
	// Find next empty line or LRU (empty line will be lru!), write-through to memory
	lruIdx := c.GetLRU(index)
	c.Contents[index][lruIdx] = CacheLine{Valid: true, Tag: tag, Data: val, LRU: 0}
	c.UpdateLRU(index, lruIdx)
	c.Memory.Write(addr, val)
	return true
}

func (c *CacheType) UpdateLRU(setIndex int, line int) {
	set := c.Contents[setIndex]
	for i := range c.Ways {

		if i == line {
			c.Contents[setIndex][i].LRU = 0
		} else if set[i].LRU < c.Ways-1 {
			c.Contents[setIndex][i].LRU += 1
		}
	}
}

func (c *CacheType) GetLRU(setIndex int) int {
	set := c.Contents[setIndex]
	lru := -1

	for i := range c.Ways {
		if set[i].LRU > lru {
			lru = i
		}
	}
	return lru
}

func (cache *CacheType) PrintCache() {

	fmt.Println("Tag    Index        Data    Valid    LRU")
	for i := range cache.Contents {
		for j := 0; j < len(cache.Contents[i]); j++ {
			line := cache.Contents[i][j]
			fmt.Printf("%05b    %03b    %08x    %t    %d\n", line.Tag, i, line.Data, line.Valid, line.LRU)
		} // might have to adjust depending on cache configs --> but nice looking for default cache
	}
	fmt.Println("DONE")
	fmt.Println("")
}
