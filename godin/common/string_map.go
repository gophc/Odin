package common

// StringHashKey bundles a string and its pre-computed hash.
type StringHashKey struct {
	Str  String
	Hash u32
}

// stringHash computes the hash of a String.
func stringHash(s String) u32 {
	h := u32(HashString(s))
	if h == 0 {
		h = 1
	}
	return h
}

// StringHash computes a StringHashKey.
func StringHash(s String) StringHashKey {
	return StringHashKey{Str: s, Hash: stringHash(s)}
}

// HashString computes a 32-bit hash of a String.
func HashString(s String) u32 {
	return FNV32a(s.Text)
}

// HashString16 computes a 32-bit hash of a String16.
func HashString16(s String16) u32 {
	h := u32(0x811c9dc5)
	for i := 0; i < s.Len; i++ {
		b := byte(s.Text[i])
		h = (h ^ u32(b)) * 0x01000193
		b = byte(s.Text[i] >> 8)
		h = (h ^ u32(b)) * 0x01000193
	}
	if h == 0 {
		h = 1
	}
	return h
}

// StringMapEntry is an entry in a StringMap.
type StringMapEntry[T any] struct {
	Key   String
	Hash  MapIndex
	Next  MapIndex
	Value T
}

// StringMap is a hash map keyed by String.
type StringMap[T any] struct {
	Hashes  []MapIndex
	Entries []StringMapEntry[T]
	Count   int
	Cap     int
}

// StringMapInit initializes a StringMap.
func StringMapInit[T any](m *StringMap[T], capacity int) {
	if capacity < 16 {
		capacity = 16
	}
	cap := nextPow2Int(capacity)
	m.Hashes = make([]MapIndex, cap)
	m.Entries = make([]StringMapEntry[T], cap)
	m.Cap = cap
	m.Count = 0
	for i := range m.Hashes {
		m.Hashes[i] = MAP_SENTINEL
	}
}

// StringMapDestroy frees the map.
func StringMapDestroy[T any](m *StringMap[T]) {
	m.Hashes = nil
	m.Entries = nil
	m.Count = 0
	m.Cap = 0
}

func (m *StringMap[T]) resizeHashes(count int) {
	if count > len(m.Hashes) {
		m.Hashes = make([]MapIndex, count)
		for i := range m.Hashes {
			m.Hashes[i] = MAP_SENTINEL
		}
	}
}

func (m *StringMap[T]) reserveEntries(capacity int) {
	if capacity > len(m.Entries) {
		m.Entries = make([]StringMapEntry[T], capacity)
	}
}

func stringMapFull[T any](m *StringMap[T]) bool {
	return m.Count*2 >= m.Cap
}

func (m *StringMap[T]) addEntry(hash u32, key String) MapIndex {
	if m.Count >= len(m.Entries) {
		// Should not happen: caller ensures capacity.
		m.grow()
	}
	idx := MapIndex(m.Count)
	m.Count++
	hashIdx := MapIndex(hash & u32(m.Cap-1))
	m.Entries[idx] = StringMapEntry[T]{
		Key:   key,
		Hash:  MapIndex(hash),
		Next:  m.Hashes[hashIdx],
		Value: *new(T),
	}
	m.Hashes[hashIdx] = idx
	return idx
}

// StringMapGet retrieves a value by String key.
func StringMapGet[T any](m *StringMap[T], key String) *T {
	if m.Count == 0 {
		return nil
	}
	h := stringHash(key)
	hashIdx := MapIndex(h & u32(m.Cap-1))
	for idx := m.Hashes[hashIdx]; idx != MAP_SENTINEL; idx = m.Entries[idx].Next {
		entry := &m.Entries[idx]
		if entry.Hash == MapIndex(h) && StringEq(entry.Key, key) {
			return &entry.Value
		}
	}
	return nil
}

func StringMapTryGet[T any](m *StringMap[T], hash u32, key String) (*T, MapFindResult) {
	fr := MapFindResult{
		HashIndex:  MAP_SENTINEL,
		EntryPrev:  MAP_SENTINEL,
		EntryIndex: MAP_SENTINEL,
	}
	if m.Count == 0 {
		fr.HashIndex = MapIndex(hash & u32(m.Cap-1))
		return nil, fr
	}
	hashIdx := MapIndex(hash & u32(m.Cap-1))
	fr.HashIndex = hashIdx
	prev := MAP_SENTINEL
	for idx := m.Hashes[hashIdx]; idx != MAP_SENTINEL; idx = m.Entries[idx].Next {
		entry := &m.Entries[idx]
		if entry.Hash == MapIndex(hash) && StringEq(entry.Key, key) {
			fr.EntryPrev = prev
			fr.EntryIndex = idx
			return &entry.Value, fr
		}
		prev = idx
	}
	return nil, fr
}

func StringMapMustGet[T any](m *StringMap[T], key String) *T {
	v := StringMapGet(m, key)
	if v == nil {
		panic("key not found in StringMap")
	}
	return v
}

// StringMapSet inserts or updates a key-value pair.
func StringMapSet[T any](m *StringMap[T], key String, value T) {
	if stringMapFull(m) {
		m.grow()
	}
	h := stringHash(key)
	hashIdx := MapIndex(h & u32(m.Cap-1))
	for idx := m.Hashes[hashIdx]; idx != MAP_SENTINEL; idx = m.Entries[idx].Next {
		entry := &m.Entries[idx]
		if entry.Hash == MapIndex(h) && StringEq(entry.Key, key) {
			entry.Value = value
			return
		}
	}
	newIdx := m.addEntry(h, key)
	m.Entries[newIdx].Value = value
}

func StringMapSetInternalFromTryGet[T any](m *StringMap[T], key String, value T, fr MapFindResult) {
	if fr.EntryIndex != MAP_SENTINEL {
		m.Entries[fr.EntryIndex].Value = value
		return
	}
	StringMapSet(m, key, value)
}

func StringMapSetIfNotPreviouslyExists[T any](m *StringMap[T], key String, value T) bool {
	if stringMapFull(m) {
		m.grow()
	}
	h := stringHash(key)
	hashIdx := MapIndex(h & u32(m.Cap-1))
	for idx := m.Hashes[hashIdx]; idx != MAP_SENTINEL; idx = m.Entries[idx].Next {
		entry := &m.Entries[idx]
		if entry.Hash == MapIndex(h) && StringEq(entry.Key, key) {
			return false
		}
	}
	newIdx := m.addEntry(h, key)
	m.Entries[newIdx].Value = value
	return true
}

func StringMapRemove[T any](m *StringMap[T], key String) {
	h := stringHash(key)
	hashIdx := MapIndex(h & u32(m.Cap-1))
	prev := MAP_SENTINEL
	for idx := m.Hashes[hashIdx]; idx != MAP_SENTINEL; idx = m.Entries[idx].Next {
		entry := &m.Entries[idx]
		if entry.Hash == MapIndex(h) && StringEq(entry.Key, key) {
			if prev == MAP_SENTINEL {
				m.Hashes[hashIdx] = entry.Next
			} else {
				m.Entries[prev].Next = entry.Next
			}
			// Clear entry.
			entry.Key = String{}
			entry.Hash = 0
			entry.Next = MAP_SENTINEL
			var zero T
			entry.Value = zero
			m.Count--
			return
		}
		prev = idx
	}
}

func StringMapClear[T any](m *StringMap[T]) {
	m.Count = 0
	for i := range m.Hashes {
		m.Hashes[i] = MAP_SENTINEL
	}
	var zero T
	for i := range m.Entries {
		m.Entries[i].Key = String{}
		m.Entries[i].Hash = 0
		m.Entries[i].Next = MAP_SENTINEL
		m.Entries[i].Value = zero
	}
}

func StringMapReserve[T any](m *StringMap[T], cap int) {
	if cap <= m.Cap {
		return
	}
	m.rehash(nextPow2Int(cap))
}

func (m *StringMap[T]) rehash(newCount int) {
	newHashes := make([]MapIndex, newCount)
	for i := range newHashes {
		newHashes[i] = MAP_SENTINEL
	}
	for i := 0; i < m.Count; i++ {
		entry := &m.Entries[i]
		h := u32(entry.Hash)
		hashIdx := MapIndex(h & u32(newCount-1))
		entry.Next = newHashes[hashIdx]
		newHashes[hashIdx] = MapIndex(i)
	}
	m.Hashes = newHashes
	m.Cap = newCount
}

func (m *StringMap[T]) grow() {
	m.rehash(m.Cap * 2)
}

// --- Iterators ---

func StringMapBegin[T any](m *StringMap[T]) *StringMapEntry[T] {
	if m.Count == 0 {
		return nil
	}
	return &m.Entries[0]
}

func StringMapEnd[T any](m *StringMap[T]) *StringMapEntry[T] {
	if m.Count == 0 {
		return nil
	}
	return &m.Entries[m.Count]
}
