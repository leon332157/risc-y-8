package memory

type MemoryResult int32
type WriteResult MemoryResult

type ReadResult struct {
	State MemoryResult
	Value uint32
}

const (
	FAILURE_INVALID_STATE = iota
	SUCCESS
	WAIT
	WAIT_NEXT_LEVEL
	FAILURE_OUT_OF_RANGE
	FAILURE
)

type Requester int

const (
	NONE Requester = 0 // No requester, used for idle state
	FETCH Requester = -1// Fetch stage in pipeline
	MEMORY Requester = -2// Memory stage in pipeline
)

const (
	LAST_LEVEL_CACHE Requester = L1_CACHE 
	L1_CACHE Requester = 1
	L2_CACHE Requester = 2
)

type Memory interface {
	IsBusy() bool // returns if memory is busy
	service(who Requester) bool // returns if memory can service a new request, but also update state
	Read(addr uint, who Requester) ReadResult
	Write(addr uint, who Requester, val uint32) WriteResult
	SizeBytes() uint // Returns the size of the memory in bytes
	SizeWords() uint // Returns the number of words in memory
	SizeLines() uint // Returns the number of lines in the memory
	RequestState() MemoryRequestState // Returns the current state of the memory request
}

type MemoryRequestState struct {
	//busy bool
	requester Requester // Who is requesting the memory service (FETCH, MEMORY, CACHE)
	Delay uint
	CyclesLeft int
}