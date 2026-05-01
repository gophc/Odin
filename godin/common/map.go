package common

import "unsafe"

// PtrMapEntry is a key-value entry in a pointer-keyed hash map.
type PtrMapEntry[K, V any] struct {
	key   K
	value V
}

// PtrMap is a hash map keyed by pointer-sized keys (integers or pointers).
type PtrMap[K, V any] struct {
	entries  []PtrMapEntry[K, V]
	count    int
	capacity int
}

func ptrMapHashKey(key uintptr) u32 {
	return FNV32a(unsafe.Slice((*byte)(unsafe.Pointer(&key)), unsafe.Sizeof(key)))
}

// PtrMapInit initializes a PtrMap with the given capacity.
func PtrMapInit[K, V any](m *PtrMap[K, V], capacity int) {
	if capacity < 16 {
		capacity = 16
	}
	capacity = nextPow2Int(capacity)
	*m = PtrMap[K, V]{
		entries:  make([]PtrMapEntry[K, V], capacity),
		capacity: capacity,
	}
}

// PtrMapDestroy frees the map.
func PtrMapDestroy[K, V any](m *PtrMap[K, V]) {
	m.entries = nil
	m.count = 0
	m.capacity = 0
}

// PtrMapGet retrieves a value by key. Returns nil if not found.
func PtrMapGet[K, V any](m *PtrMap[K, V], key K) *V {
	if m.count == 0 || m.capacity == 0 {
		return nil
	}
	h := ptrMapHashKey(uintptr(unsafe.Pointer(&key)))
	idx := int(h & u32(m.capacity-1))
	for i := 0; i < m.capacity; i++ {
		entry := &m.entries[idx]
		// Compare pointer-sized key via unsafe.
		if unsafeEqual(unsafe.Pointer(&entry.key), unsafe.Pointer(&key), unsafe.Sizeof(key)) {
			return &entry.value
		}
		if isZeroValue(unsafe.Pointer(&entry.key), unsafe.Sizeof(key)) {
			return nil
		}
		idx = (idx + 1) & (m.capacity - 1)
	}
	return nil
}

// PtrMapTryGet retrieves a value and the found index.
func PtrMapTryGet[K, V any](m *PtrMap[K, V], key K) (*V, MapIndex) {
	if m.count == 0 || m.capacity == 0 {
		return nil, MAP_SENTINEL
	}
	h := ptrMapHashKey(uintptr(unsafe.Pointer(&key)))
	idx := int(h & u32(m.capacity-1))
	for i := 0; i < m.capacity; i++ {
		entry := &m.entries[idx]
		if unsafeEqual(unsafe.Pointer(&entry.key), unsafe.Pointer(&key), unsafe.Sizeof(key)) {
			return &entry.value, MapIndex(idx)
		}
		if isZeroValue(unsafe.Pointer(&entry.key), unsafe.Sizeof(key)) {
			return nil, MAP_SENTINEL
		}
		idx = (idx + 1) & (m.capacity - 1)
	}
	return nil, MAP_SENTINEL
}

// PtrMapSetInternalFromTryGet sets a value using a pre-computed index.
func PtrMapSetInternalFromTryGet[K, V any](m *PtrMap[K, V], key K, value V, foundIndex MapIndex) {
	if foundIndex != MAP_SENTINEL {
		m.entries[foundIndex].value = value
		return
	}
	PtrMapSet(m, key, value)
}

// PtrMapMustGet returns a reference to the value, panicking if not found.
func PtrMapMustGet[K, V any](m *PtrMap[K, V], key K) *V {
	v := PtrMapGet(m, key)
	if v == nil {
		panic("key not found in PtrMap")
	}
	return v
}

// PtrMapSet inserts or updates a key-value pair.
func PtrMapSet[K, V any](m *PtrMap[K, V], key K, value V) {
	if m.count+1 >= m.capacity*3/4 {
		ptrMapGrow(m)
	}
	h := ptrMapHashKey(uintptr(unsafe.Pointer(&key)))
	idx := int(h & u32(m.capacity-1))
	for {
		entry := &m.entries[idx]
		if unsafeEqual(unsafe.Pointer(&entry.key), unsafe.Pointer(&key), unsafe.Sizeof(key)) {
			entry.value = value
			return
		}
		if isZeroValue(unsafe.Pointer(&entry.key), unsafe.Sizeof(key)) {
			entry.key = key
			entry.value = value
			m.count++
			return
		}
		idx = (idx + 1) & (m.capacity - 1)
	}
}

// PtrMapSetIfNotPreviouslyExists inserts only if the key does not already exist.
func PtrMapSetIfNotPreviouslyExists[K, V any](m *PtrMap[K, V], key K, value V) bool {
	if m.count+1 >= m.capacity*3/4 {
		ptrMapGrow(m)
	}
	h := ptrMapHashKey(uintptr(unsafe.Pointer(&key)))
	idx := int(h & u32(m.capacity-1))
	for {
		entry := &m.entries[idx]
		if unsafeEqual(unsafe.Pointer(&entry.key), unsafe.Pointer(&key), unsafe.Sizeof(key)) {
			return false
		}
		if isZeroValue(unsafe.Pointer(&entry.key), unsafe.Sizeof(key)) {
			entry.key = key
			entry.value = value
			m.count++
			return true
		}
		idx = (idx + 1) & (m.capacity - 1)
	}
}

// PtrMapRemove removes a key from the map.
func PtrMapRemove[K, V any](m *PtrMap[K, V], key K) {
	if m.count == 0 || m.capacity == 0 {
		return
	}
	h := ptrMapHashKey(uintptr(unsafe.Pointer(&key)))
	idx := int(h & u32(m.capacity-1))
	for i := 0; i < m.capacity; i++ {
		entry := &m.entries[idx]
		if unsafeEqual(unsafe.Pointer(&entry.key), unsafe.Pointer(&key), unsafe.Sizeof(key)) {
			// Clear the entry.
			var zeroK K
			var zeroV V
			entry.key = zeroK
			entry.value = zeroV
			m.count--
			return
		}
		if isZeroValue(unsafe.Pointer(&entry.key), unsafe.Sizeof(key)) {
			return
		}
		idx = (idx + 1) & (m.capacity - 1)
	}
}

// PtrMapClear removes all entries.
func PtrMapClear[K, V any](m *PtrMap[K, V]) {
	m.count = 0
	var zeroK K
	var zeroV V
	for i := range m.entries {
		m.entries[i].key = zeroK
		m.entries[i].value = zeroV
	}
}

func ptrMapGrow[K, V any](m *PtrMap[K, V]) {
	newCap := m.capacity * 2
	if newCap < 16 {
		newCap = 16
	}
	newEntries := make([]PtrMapEntry[K, V], newCap)
	for i := 0; i < m.capacity; i++ {
		entry := &m.entries[i]
		if !isZeroValue(unsafe.Pointer(&entry.key), unsafe.Sizeof(entry.key)) {
			h := ptrMapHashKey(uintptr(unsafe.Pointer(&entry.key)))
			idx := int(h & u32(newCap-1))
			for !isZeroValue(unsafe.Pointer(&newEntries[idx].key), unsafe.Sizeof(newEntries[idx].key)) {
				idx = (idx + 1) & (newCap - 1)
			}
			newEntries[idx] = *entry
		}
	}
	m.entries = newEntries
	m.capacity = newCap
}

func PtrMapReserve[K, V any](m *PtrMap[K, V], cap int) {
	if cap <= m.capacity {
		return
	}
	ptrMapGrow(m)
}

// --- Unsafe helpers ---

func unsafeEqual(a, b unsafe.Pointer, size uintptr) bool {
	switch size {
	case 1:
		return *(*byte)(a) == *(*byte)(b)
	case 2:
		return *(*uint16)(a) == *(*uint16)(b)
	case 4:
		return *(*uint32)(a) == *(*uint32)(b)
	case 8:
		return *(*uint64)(a) == *(*uint64)(b)
	default:
		return *(*uintptr)(a) == *(*uintptr)(b)
	}
}

func isZeroValue(ptr unsafe.Pointer, size uintptr) bool {
	switch size {
	case 1:
		return *(*byte)(ptr) == 0
	case 2:
		return *(*uint16)(ptr) == 0
	case 4:
		return *(*uint32)(ptr) == 0
	case 8:
		return *(*uint64)(ptr) == 0
	default:
		return *(*uintptr)(ptr) == 0
	}
}
