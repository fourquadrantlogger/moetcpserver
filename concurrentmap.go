package moetcpserver

import (
	"math/rand"
	"runtime"
	"sync"
)

var (
	defaultShardCount uint8 = uint8(runtime.NumCPU())
)

// syncMap wraps built-in map by using RWMutex for concurrent safe.
type syncMap struct {
	items map[string]Client_Con_io
	sync.RWMutex
}

// SyncMap keeps a slice of *syncMap with length of `shardCount`.
// Using a slice of syncMap instead of a large one is to avoid lock bottlenecks.
type SyncMap struct {
	Deletefunc func(m *SyncMap, addr string)
	Setfunc    func(addr string, value Client_Con_io)
	shardCount uint8
	shards     []*syncMap
}

// Create a new SyncMap with default shard count.
func SyncmapNew(deletefunc func(*SyncMap, string), setfunc func(string, Client_Con_io)) *SyncMap {
	return SyncmapNewWithShard(deletefunc, setfunc, defaultShardCount)
}

// Create a new SyncMap with given shard count.
// NOTE: shard count must be power of 2, default shard count will be used otherwise.
func SyncmapNewWithShard(deletefunc func(*SyncMap, string), setfunc func(string, Client_Con_io), shardCount uint8) *SyncMap {
	if !isPowerOfTwo(shardCount) {
		shardCount = defaultShardCount
	}
	m := new(SyncMap)
	m.shardCount = shardCount
	m.shards = make([]*syncMap, m.shardCount)
	m.Deletefunc = deletefunc
	m.Setfunc = setfunc
	for i, _ := range m.shards {
		m.shards[i] = &syncMap{items: make(map[string]Client_Con_io)}
	}
	return m
}

// Find the specific shard with the given key
func (m *SyncMap) locate(key string) *syncMap {
	return m.shards[bkdrHash(key)&uint32((m.shardCount-1))]
}

const seed uint32 = 131 // 31 131 1313 13131 131313 etc..

func bkdrHash(str string) uint32 {
	var h uint32

	for _, c := range str {
		h = h*seed + uint32(c)
	}

	return h
}

// Retrieves a value
func (m *SyncMap) Get(key string) (value Client_Con_io, ok bool) {
	shard := m.locate(key)
	shard.RLock()
	value, ok = shard.items[key]
	shard.RUnlock()
	return
}

// Sets value with the given key
func (m *SyncMap) Set(addr string, value Client_Con_io) {
	shard := m.locate(addr)
	shard.Lock()
	shard.items[addr] = value
	shard.Unlock()
	m.Setfunc(addr, value)
}

// Removes an item
func (m *SyncMap) Delete(key string) {
	shard := m.locate(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
	m.Deletefunc(m, key)
}

// Pop delete and return a random item in the cache
func (m *SyncMap) Pop() (string, Client_Con_io) {
	if m.Size() == 0 {
		panic("syncmap: map is empty")
	}

	var (
		key   string
		value Client_Con_io
		found = false
		n     = int(m.shardCount)
	)

	for !found {
		idx := rand.Intn(n)
		shard := m.shards[idx]
		shard.Lock()
		if len(shard.items) > 0 {
			found = true
			for key, value = range shard.items {
				break
			}
			delete(shard.items, key)
		}
		shard.Unlock()
	}

	return key, value
}

// Whether SyncMap has the given key
func (m *SyncMap) Has(key string) bool {
	_, ok := m.Get(key)
	return ok
}

// Returns the number of items
func (m *SyncMap) Size() int {
	size := 0
	for _, shard := range m.shards {
		shard.RLock()
		size += len(shard.items)
		shard.RUnlock()
	}
	return size
}

// Wipes all items from the map
func (m *SyncMap) Flush() int {
	size := 0
	for _, shard := range m.shards {
		shard.Lock()
		size += len(shard.items)
		shard.items = make(map[string]Client_Con_io)
		shard.Unlock()
	}
	return size
}

// Returns a channel from which each key in the map can be read
func (m *SyncMap) IterKeys() <-chan string {
	ch := make(chan string)
	go func() {
		for _, shard := range m.shards {
			shard.RLock()
			for key, _ := range shard.items {
				ch <- key
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// Item is a pair of key and value
type Item struct {
	Key   string
	Value Client_Con_io
}

// Return a channel from which each item (key:value pair) in the map can be read
func (m *SyncMap) IterItems() <-chan Item {
	ch := make(chan Item)
	go func() {
		for _, shard := range m.shards {
			shard.RLock()
			for key, value := range shard.items {
				ch <- Item{key, value}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

func isPowerOfTwo(x uint8) bool {
	return x != 0 && (x&(x-1) == 0)
}
