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
	firstZero := newMem.Read(0, LAST_LEVEL_CACHE)
	if firstZero.Value != 0 {
		t.Errorf("firstZero = %08x; want 0x0", firstZero)
	}
}

func TestReadRandom(t *testing.T) {
	newMem := CreateRAM(32, 8, 5)
	randomZero := newMem.Read(125, LAST_LEVEL_CACHE)
	if randomZero.Value != 0 {
		t.Errorf("firstZero = %08x; want 0x0", randomZero)
	}
}

func TestWriteFrontNoDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 0)
	newMem.Write(0, LAST_LEVEL_CACHE, 1)
	read := newMem.Read(0, LAST_LEVEL_CACHE).Value
	if read != 1 {
		t.Errorf("read resulted in %08x; want 0x1", read)
	}
}

func TestWriteEndNoDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 0)
	newMem.Write(255, LAST_LEVEL_CACHE, 28) // 0x1C in hex
	read := newMem.Read(255, LAST_LEVEL_CACHE).Value
	if read != 28 {
		t.Errorf("read resulted in %08x; want 0x1C", read)
	}
}

func TestWriteNoDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 0)
	newMem.Write(32, LAST_LEVEL_CACHE, 3735928559)
	newMem.Write(197, LAST_LEVEL_CACHE, 65535)

	read1 := newMem.Read(197, LAST_LEVEL_CACHE)
	if read1.Value != 65535 {
		t.Errorf("read resulted in %08x; want 0xFFFF", read1)
	}

	read2 := newMem.Read(32, LAST_LEVEL_CACHE)
	if read2.Value != 3735928559 {
		t.Errorf("read resulted in %08x; want 0xDEADBEEF", read2)
	}
}

func TestWrite1And9(t *testing.T) {
	mem := CreateRAM(32, 8, 0)
	mem.Write(1, LAST_LEVEL_CACHE, 0xfeebdaed)

	read1 := mem.Read(1, LAST_LEVEL_CACHE)
	read9 := mem.Read(9, LAST_LEVEL_CACHE)

	if read1.Value != 0xfeebdaed {
		t.Errorf("read 0x1 resulted in %08x; want 0xfeebdaed", read1)
	}
	if read9.Value != 0 {
		t.Errorf("read resulted in %08x; want 0x0", read9)
	}

}

func TestWriteSameLocationNoDelay(t *testing.T) {
	newMem := CreateRAM(32, 8, 0)

	newMem.Write(32, LAST_LEVEL_CACHE, 3735928559)
	read1 := newMem.Read(32, LAST_LEVEL_CACHE)
	if read1.Value != 3735928559 {
		t.Errorf("read resulted in %08x; want 0xDEADBEEF", read1)
	}

	newMem.Write(32, LAST_LEVEL_CACHE, 65535)
	read2 := newMem.Read(32, LAST_LEVEL_CACHE)
	if read2.Value != 65535 {
		t.Errorf("read resulted in %08x; want 0xFFFF", read2)
	}
}

func TestDifferentStageAccess(t *testing.T) {
	mem := CreateRAM(32, 8, 2)
	for range 3 {
		mem.Write(1, LAST_LEVEL_CACHE, 0xfeebdaed)
	}
	
	read1 := mem.Read(1, LAST_LEVEL_CACHE)
	write := mem.Write(9, LAST_LEVEL_CACHE, 123)

	if read1.State != SUCCESS {
		t.Errorf("is reading from RAM while it is servicing a different stage %d", read1)
	}

	if write != WAIT {
		t.Errorf("ram should return wait %d", write)
	}

}

func TestStagingFiveDelayWrite(t *testing.T) {
	mem := CreateRAM(32, 8, 5)

	call1 := mem.Write(1, LAST_LEVEL_CACHE, 0x122122)
	call2 := mem.Write(1, LAST_LEVEL_CACHE, 0x122122)
	call3 := mem.Write(1, LAST_LEVEL_CACHE, 0x122122)
	call4 := mem.Write(1, LAST_LEVEL_CACHE, 0x122122)
	call5 := mem.Write(1, LAST_LEVEL_CACHE, 0x122122)
	call6 := mem.Write(1, LAST_LEVEL_CACHE, 0x122122)
	call7 := mem.Write(1, LAST_LEVEL_CACHE, 0x122122)

	if call1 != WAIT {
		t.Errorf("ram should return wait, got %d", call1)
	}
	if call2 != WAIT {
		t.Errorf("ram should return wait %d", call2)
	}
	if call3 != WAIT {
		t.Errorf("ram should return wait %d", call3)
	}
	if call4 != WAIT {
		t.Errorf("ram should return wait %d", call4)
	}
	if call5 != WAIT {
		t.Errorf("ram should return wait %d", call5)
	}
	if call6 != SUCCESS {
		t.Errorf("ram should return success, got %d", call6)
	}
	if call7 != WAIT {
		t.Errorf("ram should return success, got %d", call7)
	}

}