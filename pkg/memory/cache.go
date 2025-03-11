package memory

import (
	"container/list" // for LRU queue
)

type Cache interface {
	createDefault(mem RAM) CacheType
	configureCache(lineSize, numSets, ways, latency int, mem RAM) CacheType
	search(addr int) bool
}

type CacheLine struct {
	tag   int
	data  []uint32
	valid bool
	dirty bool
}

type Set struct {
	lines    []CacheLine
	LRUQueue *list.List // Tracks LRU order with a queue
}

type CacheType struct {
	lineSize int
	numSets  int
	ways     int
	access   AccessState
	sets     []Set
	memory   RAM
}
