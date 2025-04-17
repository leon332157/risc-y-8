package memory

import (
	"testing"
)

func TestFindIndexAndTag(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := CreateCacheDefault(&newMem)

	idxTag := c.FindIndexAndTag(128)
	idx, tag := idxTag.index, idxTag.tag

	if idx != 0 {
		t.Errorf("index = %d; want 0", idx)
	}
	if tag != 0b10000 {
		t.Errorf("tag = %b; want 0b01000", tag)
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

	read = c.Read(32, FETCH_STAGE)

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

func TestWriteThrough(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := CreateCacheDefault(&newMem)

	c.Write(3, MEMORY_STAGE, 0x123456)
	readMem := newMem.Read(3, LAST_LEVEL_CACHE)
	readC := c.Read(3, MEMORY_STAGE)

	if readMem.Value != 0x123456 && readC.Value != 0x123456 {
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
		t.Errorf("should be wait on mem, got %d", call1)
	}
	if call2.State != WAIT_NEXT_LEVEL {
		t.Errorf("should be wait on mem, got %d", call2)
	}
	if call3.State != WAIT {
		t.Errorf("should be wait, got %d", call3)
	}
	if call4.State != SUCCESS {
		t.Errorf("should be success, got %d", call4)
	}
}

func TestCacheLoadsLineNoDelay(t *testing.T) {
	mem := CreateRAM(32, 8, 0)
	c := CreateCache(8, 2, 4, 0, &mem)

	mem.Write()
}
