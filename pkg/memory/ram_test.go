package memory

import (
	"testing"
)

func TestCreate(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	size := len(newMem.Contents)
	if size != 256 {
		t.Errorf("mem size = %d; want 256", size)
	}
}

func TestReadEmpty(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	firstZero := newMem.Read(0)
	if firstZero != 0 {
		t.Errorf("firstZero = %08x; want 0x0", firstZero)
	}
}

func TestReadRandom(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	randomZero := newMem.Read(125)
	if randomZero != 0 {
		t.Errorf("firstZero = %08x; want 0x0", randomZero)
	}
}

func TestWriteFrontNoDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 0)
	newMem.Write(0, 1)
	read := newMem.Read(0)
	if read != 1 {
		t.Errorf("read resulted in %08x; want 0x1", read)
	}
}

func TestWriteEndNoDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 0)
	newMem.Write(255, 28)
	read := newMem.Read(255)
	if read != 28 {
		t.Errorf("read resulted in %08x; want 0x1C", read)
	}
}

func TestWriteNoDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 0)
	newMem.Write(32, 3735928559)
	newMem.Write(197, 65535)

	read1 := newMem.Read(197)
	if read1 != 65535 {
		t.Errorf("read resulted in %08x; want 0xFFFF", read1)
	}

	read2 := newMem.Read(32)
	if read2 != 3735928559 {
		t.Errorf("read resulted in %08x; want 0xDEADBEEF", read2)
	}
}

func TestWriteSameLocationNoDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 0)

	newMem.Write(32, 3735928559)
	read1 := newMem.Read(32)
	if read1 != 3735928559 {
		t.Errorf("read resulted in %08x; want 0xDEADBEEF", read1)
	}

	newMem.Write(32, 65535)
	read2 := newMem.Read(32)
	if read2 != 65535 {
		t.Errorf("read resulted in %08x; want 0xFFFF", read2)
	}
}
