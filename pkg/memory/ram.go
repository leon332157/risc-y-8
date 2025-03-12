package memory

import "fmt"

// Create a memory interface
type Memory interface {
    CreateRAM(size, lineSize, wordSize, latency int) RAM
    Read(addr int, lin bool) *RAMValue
    Write(addr int, val *RAMValue) bool
}

// A line of memory or value in memory
type RAMValue struct {
    Line  []uint32
    value uint32
}

// RAM type with size and memory attributes
type RAM struct {
    size            int
    lineSize        int
    wordSize        int
    Access          AccessState
    Contents        [][]uint32
    WriteCyclesLeft int
    WriteInProgress bool
    WriteAddr       int
    WriteData       []uint32
    LastReadAddr    int // Stores the last read address
    ReadInProgress  bool // Indicates if a read is ongoing
}

// AccessState tracks memory access
type AccessState struct {
    latency    int
    CyclesLeft int
    accessed   bool
}

// AccessControl constructor
func createAccessState(latency int) *AccessState {
    return &AccessState{
        latency:    latency,
        CyclesLeft: latency,
        accessed:   false,
    }
}

// Returns whether memory can be accessed
func (c *AccessState) AccessAttempt() bool {
    if c.accessed && c.CyclesLeft != 0 {
        c.CyclesLeft--
        return false
    }
    c.accessed = true
    c.CyclesLeft = c.latency + 1
    return true
}

// RAM constructor
func CreateRAM(size, lineSize, wordSize, latency int) RAM {
    contents := make([][]uint32, size)
    for i := 0; i < size; i++ {
        contents[i] = make([]uint32, lineSize)
    }
    access := createAccessState(latency)

    return RAM{
        size:     size,
        lineSize: lineSize,
        wordSize: wordSize,
        Access:   *access,
        Contents: contents,
    }
}

// Aligns address to memory constraints
func (mem *RAM) alignRAM(addr int) int {
    return ((addr % mem.size) / mem.wordSize) * mem.wordSize
}

// Calculates block and offset from address
func (mem *RAM) addrToOffset(addr int) (int, int) {
    alignedAddr := mem.alignRAM(addr)
    return (alignedAddr / mem.wordSize) % mem.size / mem.lineSize, (alignedAddr / mem.wordSize) % mem.lineSize
}

// Reads a value from memory
func (mem *RAM) Read(addr int, lin bool) *RAMValue {

    if mem.WriteInProgress {
        mem.WriteCyclesLeft--
        fmt.Println("WAIT, memory write in progress. Cycles left:", mem.WriteCyclesLeft)
        return nil
    } else if !mem.Access.AccessAttempt() {
        fmt.Println("WAIT, memory cannot be accessed this cycle. Cycles left:", mem.Access.CyclesLeft)
        mem.ReadInProgress = true
        mem.LastReadAddr = addr
        return nil
    }

    mem.Access.accessed = false
    mem.ReadInProgress = false

    index, offset := mem.addrToOffset(addr)

    if lin {
        return &RAMValue{Line: append([]uint32{}, mem.Contents[index]...)}
    }

    return &RAMValue{value: mem.Contents[index][offset]}
}

// Writes a value to memory
func (mem *RAM) Write(addr int, val *RAMValue) bool {
    
    if mem.ReadInProgress {
        mem.Access.CyclesLeft--
        fmt.Println("WAIT, memory read in progress. Cycles left:", mem.Access.CyclesLeft)
        return false
    }

    if mem.WriteInProgress {
        if mem.WriteCyclesLeft > 0 {
            mem.WriteCyclesLeft--
            fmt.Println("WAIT, memory write in progress. Cycles left:", mem.WriteCyclesLeft)
            return false
        }
        mem.WriteInProgress = false
    }

    if !mem.WriteInProgress {
        mem.Access.AccessAttempt()
        mem.WriteInProgress = true
        mem.WriteCyclesLeft = mem.Access.latency
        mem.WriteAddr = addr
        mem.WriteData = append([]uint32{}, val.Line...)
        fmt.Println("Memory write initiated, will complete in", mem.WriteCyclesLeft, "cycles.")
        return false
    }

    if len(val.Line) > 0 {
        mem.Contents[mem.WriteAddr] = val.Line
    }

    // mem.writeInProgress = false
    fmt.Printf("Memory write completed, wrote %08X to address %d\n", val.Line, mem.WriteAddr)
    return true
}

func (m *RAM) StartRead(addr int) {
    if m.ReadInProgress {
        fmt.Println("Memory read already in progress. Wait.")
        return
    }

    m.ReadInProgress = true
    m.Access.CyclesLeft = m.Access.latency
    m.LastReadAddr = addr
    fmt.Println("Memory read started. Will take", m.Access.latency, "cycles.")
}


// Prints memory in hex
func Print2DSlice(slice [][]uint32) {
    for _, row := range slice {
        for _, val := range row {
            fmt.Printf("0x%08X ", val)
        }
        fmt.Println()
    }
}