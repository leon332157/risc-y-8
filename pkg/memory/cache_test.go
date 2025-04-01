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
		newMem.Write(32, LAST_LEVEL_CACHE,0xDEADBEEF)
	}

	var read ReadResult

	read = c.Read(32, FETCH)

	if read.Value != 3735928559 {
		t.Errorf("read 1 resulted in %08x; want 0xDEADBEEF", read.Value)
	}

	// Check if loaded into cache, read hit
	newMem.Write(32,LAST_LEVEL_CACHE, 0xFFFF)
	read = c.Read(32, FETCH)

	if read.Value != 3735928559 {
		t.Errorf("read 2 resulted in %08x; want 0xDEADBEEF", read.Value)
	}

}

func TestWriteThrough(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := CreateCacheDefault(&newMem)

	c.Write(3, MEMORY, 0x123456)
	readMem := newMem.Read(3,LAST_LEVEL_CACHE)
	readC := c.Read(3, MEMORY)

	if readMem.Value != 0x123456 && readC.Value != 0x123456 {
		t.Errorf("mem read resulted in %08x; want 0x123456", readMem.Value)
		t.Errorf("cache read resulted in %08x; want 0x123456", readC.Value)
	}
}