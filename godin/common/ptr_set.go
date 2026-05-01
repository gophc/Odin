package common

import "unsafe"

// PtrSet is a hash set for pointer-sized keys (integers, pointers).
type PtrSet[T comparable] struct {
	keys     []T
	count    int
	capacity int
	inline   [PTR_SET_INLINE_CAP]T
}

const PTR_SET_INLINE_CAP = 16

func ptrSetHashKey(key uintptr) u32 {
	return FNV32a(unsafe.Slice((*byte)(unsafe.Pointer(&key)), unsafe.Sizeof(key)))
}

// PtrSetInit initializes a PtrSet.
func PtrSetInit[T comparable](s *PtrSet[T], capacity int) {
	if capacity < PTR_SET_INLINE_CAP {
		capacity = PTR_SET_INLINE_CAP
	}
	if capacity > PTR_SET_INLINE_CAP {
		s.keys = make([]T, capacity)
	} else {
		s.keys = s.inline[:]
	}
	s.capacity = capacity
	s.count = 0
}

// PtrSetDestroy frees the set.
func PtrSetDestroy[T comparable](s *PtrSet[T]) {
	s.keys = nil
	s.count = 0
	s.capacity = 0
}

// PtrSetExists checks if a pointer exists in the set.
func PtrSetExists[T comparable](s *PtrSet[T], ptr T) bool {
	if s.count == 0 || s.capacity == 0 {
		return false
	}
	h := ptrSetHashKey(uintptr(unsafe.Pointer(&ptr)))
	idx := int(h & u32(s.capacity-1))
	for i := 0; i < s.capacity; i++ {
		if s.keys[idx] == ptr {
			return true
		}
		var zero T
		if s.keys[idx] == zero {
			return false
		}
		idx = (idx + 1) & (s.capacity - 1)
	}
	return false
}

// PtrSetUpdate adds the pointer if not present. Returns true if added.
func PtrSetUpdate[T comparable](s *PtrSet[T], ptr T) bool {
	if s.count+1 >= s.capacity*3/4 {
		ptrSetGrow(s)
	}
	h := ptrSetHashKey(uintptr(unsafe.Pointer(&ptr)))
	idx := int(h & u32(s.capacity-1))
	var zero T
	for {
		if s.keys[idx] == ptr {
			return false
		}
		if s.keys[idx] == zero {
			s.keys[idx] = ptr
			s.count++
			return true
		}
		idx = (idx + 1) & (s.capacity - 1)
	}
}

// PtrSetAdd adds the pointer (no-op if exists).
func PtrSetAdd[T comparable](s *PtrSet[T], ptr T) {
	PtrSetUpdate(s, ptr)
}

// PtrSetRemove removes a pointer from the set.
func PtrSetRemove[T comparable](s *PtrSet[T], ptr T) {
	if s.count == 0 || s.capacity == 0 {
		return
	}
	h := ptrSetHashKey(uintptr(unsafe.Pointer(&ptr)))
	idx := int(h & u32(s.capacity-1))
	var zero T
	for i := 0; i < s.capacity; i++ {
		if s.keys[idx] == ptr {
			s.keys[idx] = zero
			s.count--
			return
		}
		if s.keys[idx] == zero {
			return
		}
		idx = (idx + 1) & (s.capacity - 1)
	}
}

// PtrSetClear removes all entries.
func PtrSetClear[T comparable](s *PtrSet[T]) {
	var zero T
	for i := range s.keys {
		s.keys[i] = zero
	}
	s.count = 0
}

func ptrSetGrow[T comparable](s *PtrSet[T]) {
	newCap := s.capacity * 2
	if newCap < 16 {
		newCap = 16
	}
	newKeys := make([]T, newCap)
	var zero T
	for i := 0; i < s.capacity; i++ {
		key := s.keys[i]
		if key != zero {
			h := ptrSetHashKey(uintptr(unsafe.Pointer(&key)))
			idx := int(h & u32(newCap-1))
			for newKeys[idx] != zero {
				idx = (idx + 1) & (newCap - 1)
			}
			newKeys[idx] = key
		}
	}
	s.keys = newKeys
	s.capacity = newCap
}
