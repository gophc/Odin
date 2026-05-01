package common

// StringSetEntry is an entry in a StringSet.
type StringSetEntry struct {
	Hash  MapIndex
	Next  MapIndex
	Value String
}

// StringSet is a hash set for Strings.
type StringSet struct {
	Hashes  Slice[MapIndex]
	Entries Array[StringSetEntry]
}

// StringSetInit initializes a StringSet.
func StringSetInit(s *StringSet, capacity int) {
	if capacity < 16 {
		capacity = 16
	}
	capacity = nextPow2Int(capacity)
	SliceInit(&s.Hashes, capacity)
	ArrayInitWithCountAndCapacity(&s.Entries, 0, capacity)
	for i := 0; i < capacity; i++ {
		s.Hashes.data[i] = MAP_SENTINEL
	}
}

// StringSetDestroy frees the set.
func StringSetDestroy(s *StringSet) {
	SliceFree(&s.Hashes)
	ArrayFree(&s.Entries)
}

func stringSetFull(s *StringSet) bool {
	return s.Entries.count >= s.Hashes.count
}

func (s *StringSet) addEntry(key StringHashKey) MapIndex {
	idx := MapIndex(s.Entries.count)
	e := StringSetEntry{
		Hash:  MapIndex(key.Hash),
		Next:  s.Hashes.data[key.Hash&u32(s.Hashes.count-1)],
		Value: key.Str,
	}
	s.Entries.Add(e)
	s.Hashes.data[key.Hash&u32(s.Hashes.count-1)] = idx
	return idx
}

func (s *StringSet) find(key StringHashKey) MapFindResult {
	fr := MapFindResult{
		HashIndex:  MAP_SENTINEL,
		EntryPrev:  MAP_SENTINEL,
		EntryIndex: MAP_SENTINEL,
	}
	if s.Hashes.count == 0 {
		return fr
	}
	hashIdx := MapIndex(key.Hash & u32(s.Hashes.count-1))
	fr.HashIndex = hashIdx
	prev := MAP_SENTINEL
	for idx := s.Hashes.data[hashIdx]; idx != MAP_SENTINEL; idx = s.Entries.data[idx].Next {
		entry := &s.Entries.data[idx]
		if entry.Hash == MapIndex(key.Hash) && StringEq(entry.Value, key.Str) {
			fr.EntryPrev = prev
			fr.EntryIndex = idx
			return fr
		}
		prev = idx
	}
	return fr
}

// StringSetExists checks if a string is in the set.
func StringSetExists(s *StringSet, str String) bool {
	if s.Hashes.count == 0 {
		return false
	}
	key := StringHash(str)
	fr := s.find(key)
	return fr.EntryIndex != MAP_SENTINEL
}

// StringSetAdd adds a string to the set (no-op if exists).
func StringSetAdd(s *StringSet, str String) {
	if stringSetFull(s) {
		stringSetGrow(s)
	}
	key := StringHash(str)
	fr := s.find(key)
	if fr.EntryIndex == MAP_SENTINEL {
		s.addEntry(key)
	}
}

// StringSetUpdate adds the string if not present. Returns true if added.
func StringSetUpdate(s *StringSet, str String) bool {
	if stringSetFull(s) {
		stringSetGrow(s)
	}
	key := StringHash(str)
	fr := s.find(key)
	if fr.EntryIndex == MAP_SENTINEL {
		s.addEntry(key)
		return true
	}
	return false
}

// StringSetRemove removes a string from the set.
func StringSetRemove(s *StringSet, str String) {
	if s.Hashes.count == 0 {
		return
	}
	key := StringHash(str)
	fr := s.find(key)
	if fr.EntryIndex == MAP_SENTINEL {
		return
	}
	// Remove from linked list.
	entry := &s.Entries.data[fr.EntryIndex]
	if fr.EntryPrev == MAP_SENTINEL {
		s.Hashes.data[fr.HashIndex] = entry.Next
	} else {
		s.Entries.data[fr.EntryPrev].Next = entry.Next
	}
	// Erase entry.
	entry.Hash = 0
	entry.Next = MAP_SENTINEL
	entry.Value = String{}
}

// StringSetClear removes all entries.
func StringSetClear(s *StringSet) {
	s.Entries.Clear()
	for i := 0; i < s.Hashes.count; i++ {
		s.Hashes.data[i] = MAP_SENTINEL
	}
}

// StringSetReserve ensures capacity for at least cap entries.
func StringSetReserve(s *StringSet, cap int) {
	if cap <= s.Hashes.count {
		return
	}
	stringSetRehash(s, nextPow2Int(cap))
}

func stringSetRehash(s *StringSet, newCount int) {
	s.Hashes.Resize(newCount)
	for i := 0; i < newCount; i++ {
		s.Hashes.data[i] = MAP_SENTINEL
	}
	for i := 0; i < s.Entries.count; i++ {
		entry := &s.Entries.data[i]
		h := u32(entry.Hash) & u32(s.Hashes.count-1)
		entry.Next = s.Hashes.data[h]
		s.Hashes.data[h] = MapIndex(i)
	}
}

func stringSetGrow(s *StringSet) {
	newCount := s.Hashes.count * 2
	if newCount < 16 {
		newCount = 16
	}
	stringSetRehash(s, newCount)
}

// --- Iterators ---

func StringSetBegin(s *StringSet) *StringSetEntry {
	if s.Entries.count == 0 {
		return nil
	}
	return &s.Entries.data[0]
}

func StringSetEnd(s *StringSet) *StringSetEntry {
	if s.Entries.count == 0 {
		return nil
	}
	return &s.Entries.data[s.Entries.count]
}
