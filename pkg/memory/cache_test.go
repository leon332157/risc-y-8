package memory

import (
	"testing"
)

func TestFindIndexAndTag(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := Default(&newMem)

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
	c := Default(&newMem)
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
	c := Default(&newMem)

	//load into mem, read miss empty cache
	newMem.Write(32, 0xDEADBEEF)
	read := c.Read(32)

	if read != 3735928559 {
		t.Errorf("read resulted in %08x; want 0xDEADBEEF", read)
	}

	// Check if loaded into cache, read hit
	newMem.Write(32, 0xFFFF)
	read = c.Read(32)

	if read != 3735928559 {
		t.Errorf("read resulted in %08x; want 0xDEADBEEF", read)
	}

}

func TestWriteThrough(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	c := Default(&newMem)

	c.Write(3, 0x123456)
	readMem := newMem.Read(3)
	readC := c.Read(3)

	if readMem != 0x123456 && readC != 0x123456 {
		t.Errorf("mem read resulted in %08x; want 0x123456", readMem)
		t.Errorf("cache read resulted in %08x; want 0x123456", readC)
	}
}
