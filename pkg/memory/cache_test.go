package memory

import (
	"testing"
)

func TestFindIndexAndTagAndOffset(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := CreateCacheDefault(&newMem) // 8 sets, 2 ways, 4 wpl 

	idxTag := c.FindIndexAndTag(2)
	idx, tag, offset := idxTag.index, idxTag.tag, idxTag.offset

	// addr = 0b0000010
	// 4 wpl --> 2 bits for offset
	// 8 sets --> 3 bits for index
	if idx != 0b000 {
		t.Errorf("index = %b; want 000", idx)
	}
	if offset != 0b10 {
		t.Errorf("offset = %d; want 10", offset)
	}
	if tag != 0 {
		t.Errorf("tag = %b; want 0", tag)
	}
}

func TestCreateCacheDefault(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := CreateCacheDefault(&newMem)
	sets := len(c.Contents)
	ways := len(c.Contents[7])

	if sets != 8 {
		t.Errorf("num sets = %d; want 8", sets)
	}

	if ways != 2 {
		t.Errorf("num ways = %d; want 2", ways)
	}
}

func TestReadHitAndMiss(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := CreateCacheDefault(&newMem)

	//load into mem, read miss empty cache
	for range 6 {
		newMem.Write(32, LAST_LEVEL_CACHE, 0xDEADBEEF)
	}

	var read ReadResult

	for range 6 {
		read = c.Read(32, FETCH_STAGE)
	}

	if read.Value != 3735928559 {
		t.Errorf("read 1 resulted in %08x; want 0xDEADBEEF", read.Value)
	}

	// Check if loaded into cache, read hit
	newMem.Write(32, LAST_LEVEL_CACHE, 0xFFFF)
	read = c.Read(32, FETCH_STAGE)

	if read.Value != 3735928559 {
		t.Errorf("read 2 resulted in %08x; want 0xDEADBEEF", read.Value)
	}

}

func TestCacheHit(t *testing.T) {

	newMem := CreateRAM(32, 8, 5)
	c := CreateCacheDefault(&newMem)

	// writes to cache and memory
	for range c.MemoryRequestState.Delay + 1{
		c.Write(32, LAST_LEVEL_CACHE, 0xDEADBEEF)
	}

	for range c.MemoryRequestState.Delay + 1{
		c.Write(33, LAST_LEVEL_CACHE, 0xCAFEBABE)
	}

	// if 0 cycle read, it was a hit
	read := c.Read(32, FETCH_STAGE)

	if read.State != SUCCESS || c.Contents[0][0].Data[0] != 0xDEADBEEF || c.Contents[0][0].Data[1] != 0xCAFEBABE {
		t.Errorf("read 1 resulted in %08x; want 0xDEADBEEF", read.Value)
	}

}

func TestCacheHitDelay(t *testing.T) {

	newMem := CreateRAM(32, 8, 5)
	c := CreateCache(8, 2, 4, 5, &newMem)

	// writes to cache and memory
	for range c.MemoryRequestState.Delay + 1{
		c.Write(32, LAST_LEVEL_CACHE, 0xDEADBEEF)
	}

	// if 0 cycle read, it was a hit
	var read ReadResult
	
	for range 6 {
		read = c.Read(32, FETCH_STAGE)
	}

	if read.State != SUCCESS {
		t.Errorf("read 1 resulted in %08x; want 0xDEADBEEF", read.Value)
	}

}

func TestWriteThrough(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := CreateCacheDefault(&newMem)

	for range 6 {
		c.Write(3, MEMORY_STAGE, 0x123456)
	}
	for range 5 {
		newMem.Read(3, LAST_LEVEL_CACHE)
	}
	readMem := newMem.Read(3, LAST_LEVEL_CACHE)
	readC := c.Read(3, MEMORY_STAGE)

	if readMem.Value != 0x123456 || readC.Value != 0x123456 {
		t.Errorf("mem read resulted in %08x; want 0x123456", readMem.Value)
		t.Errorf("cache read resulted in %08x; want 0x123456", readC.Value)
	}
}

func TestStagingDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := CreateCache(8, 2, 4, 1, &newMem)

	c.Write(3, FETCH_STAGE, 0xFFFFFF)

}

func TestStagingReadDelay(t *testing.T) {
	mem := CreateRAM(32, 8, 2)
	c := CreateCache(8, 2, 4, 2, &mem)

	call1 := c.Read(1, FETCH_STAGE)
	call2 := c.Read(1, FETCH_STAGE)
	call3 := c.Read(1, MEMORY_STAGE)
	call4 := c.Read(1, FETCH_STAGE)

	if call1.State != WAIT_NEXT_LEVEL {
		t.Errorf("should be wait on mem, got %d", call1.State)
	}
	if call2.State != WAIT_NEXT_LEVEL {
		t.Errorf("should be wait on mem, got %d", call2.State)
	}
	if call3.State != WAIT {
		t.Errorf("should be wait, got %d", call3.State)
	}
	if call4.State != SUCCESS {
		t.Errorf("should be success, got %d", call4.State)
	}

}

func TestCacheLoadsLineNoDelay(t *testing.T) {
	mem := CreateRAM(32, 8, 0)
	c := CreateCache(8, 2, 4, 0, &mem)

	mem.Contents[0] = 0xffffff
	mem.Contents[1] = 0xdead00
	mem.Contents[2] = 0x123456
	mem.Contents[3] = 0x765432
	read1 := c.Read(1, MEMORY_STAGE)
	read2 := c.Read(3, MEMORY_STAGE)
	contents := c.Contents[0][0].Data[2]

	if read2.State != SUCCESS {
		t.Errorf("should be success")
	}
	if read1.Value != 0xdead00 {
		t.Errorf("should be 0xdead00, got %08X", read1.Value)
	}
	if read2.Value != 0x765432 {
		t.Errorf("should be 0x765432, got %08X", read2.Value)
	}
	if contents != 0x123456 {
		t.Errorf("should be 0x123456, got %08X", contents)
	}

}

func TestLRUUpdate(t *testing.T) {
	mem := CreateRAM(32, 8, 0)
	c := CreateCache(4, 4, 4, 0, &mem)

	mem.Write(0, LAST_LEVEL_CACHE, 0x099900)
	// pretend use set0, line 2
	set0 := c.Contents[0]

	c.Read(0, MEMORY_STAGE)

	line0_lru := set0[0].LRU // 3 -> 0
	line1_lru := set0[1].LRU // 2 -> 3
	line2_lru := set0[2].LRU // 1 -> 2
	line3_lru := set0[3].LRU // 0 -> 1

	if line0_lru != 0 || line1_lru != 3 || line2_lru != 2 || line3_lru != 1 {
		t.Errorf("new load to line 0 should be 0, got %d", line0_lru)
		t.Errorf("line 1 should update to 3, got %d", line1_lru)
		t.Errorf("line 2 should update to 2, got %d", line2_lru)
		t.Errorf("line 3 should update to 1, got %d", line3_lru)

	}
}

func TestCacheReadWithDelay(t *testing.T) {

	// 2 way, 2 set, 2 word per line, 10 cycle delay
	mem := CreateRAM(512, 8, 15)
	cache := CreateCache(2, 2, 2, 7, &mem)

	for range mem.MemoryRequestState.Delay + 1 {
		cache.Write(0, MEMORY_STAGE, 0x00112233)
	}

	for range mem.MemoryRequestState.Delay + 1 {
		cache.Write(1, MEMORY_STAGE, 0x44556677)
	}

	for range mem.MemoryRequestState.Delay + 1 {
		cache.Write(2, MEMORY_STAGE, 0x8899AABB)
	}

	for range mem.MemoryRequestState.Delay + 1 {
		cache.Write(3, MEMORY_STAGE, 0xCCDDEEFF)
	}

	if mem.Contents[0] != 0x00112233 || mem.Contents[1] != 0x44556677 ||
		mem.Contents[2] != 0x8899AABB || mem.Contents[3] != 0xCCDDEEFF {
		t.Errorf("memory contents not written correctly: %08X, %08X, %08X, %08X", mem.Contents[0], mem.Contents[1], mem.Contents[2], mem.Contents[3])
	}

	for range cache.MemoryRequestState.Delay + 1 {
		cache.Read(0, MEMORY_STAGE)
	}

	for range cache.MemoryRequestState.Delay + 1 {
		cache.Read(1, MEMORY_STAGE)
	}

	for range cache.MemoryRequestState.Delay + 1 {
		cache.Read(2, MEMORY_STAGE)
	}

	for range cache.MemoryRequestState.Delay + 1 {
		cache.Read(3, MEMORY_STAGE)
	}

	if cache.Contents[0][0].Data[0] != 0x00112233 ||
		cache.Contents[0][0].Data[1] != 0x44556677 ||
		cache.Contents[1][0].Data[0] != 0x8899AABB ||
		cache.Contents[1][0].Data[1] != 0xCCDDEEFF {
		t.Errorf("cache contents not correct")
	}

}

func TestCacheReadValuesWithDelay(t *testing.T) {

    // 2 way, 2 set, 2 word per line, 10 cycle delay
    mem := CreateRAM(512, 8, 15)
    cache := CreateCache(2, 2, 2, 7, &mem)

    for range mem.MemoryRequestState.Delay + 1 {
        cache.Write(0, MEMORY_STAGE, 0x00112233)
    }

    for range mem.MemoryRequestState.Delay + 1 {
        cache.Write(1, MEMORY_STAGE, 0x44556677)
    }

    for range mem.MemoryRequestState.Delay + 1 {
        cache.Write(2, MEMORY_STAGE, 0x8899AABB)
    }

    for range mem.MemoryRequestState.Delay + 1 {
        cache.Write(3, MEMORY_STAGE, 0xCCDDEEFF)
    }

    if mem.Contents[0] != 0x00112233 || mem.Contents[1] != 0x44556677 ||
        mem.Contents[2] != 0x8899AABB || mem.Contents[3] != 0xCCDDEEFF {
        t.Errorf("memory contents not written correctly: %08X, %08X, %08X, %08X", mem.Contents[0], mem.Contents[1], mem.Contents[2], mem.Contents[3])
    }

	var read1 ReadResult
	var read2 ReadResult
	var read3 ReadResult
	var read4 ReadResult

    for range cache.MemoryRequestState.Delay + 1 {
        read1 = cache.Read(0, MEMORY_STAGE)
    }

    for range cache.MemoryRequestState.Delay + 1 {
        read2 = cache.Read(1, MEMORY_STAGE)
    }

    for range cache.MemoryRequestState.Delay + 1 {
        read3 = cache.Read(2, MEMORY_STAGE)
    }

    for range cache.MemoryRequestState.Delay + 1 {
        read4 = cache.Read(3, MEMORY_STAGE)
    }

    if read1.Value != 0x00112233 || read2.Value != 0x44556677 || read3.Value != 0x8899AABB || read4.Value != 0xCCDDEEFF {
		t.Errorf("values are wrong")
	}

}

func TestInitLRUMultiWays(t *testing.T) {
	mem := CreateRAM(32, 8, 0)
	c := CreateCache(4, 4, 4, 0, &mem)

	line1_lru := c.Contents[0][0].LRU
	line2_lru := c.Contents[0][1].LRU
	line3_lru := c.Contents[0][2].LRU

	if line1_lru != 3 || line2_lru != 2 || line3_lru != 1 {
		t.Errorf("line 1 lru should be 2, got %d", line1_lru)
		t.Errorf("line 2 lru should be 1, got %d", line2_lru)
		t.Errorf("line 3 lru should be 0, got %d", line3_lru)
	}

	line4_lru := c.Contents[0][3].LRU

	if line4_lru != 0 {
		t.Errorf("empty line 4 should be 3, go %d", line4_lru)
	}
}
