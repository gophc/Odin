package common

// Array is a dynamic array (like C++ std::vector or Odin's array).
type Array[T any] struct {
	data     []T
	count    int
	capacity int
}

// Slice is a view into an Array or another Slice.
type Slice[T any] struct {
	data  []T
	count int
}

// --- Array functions ---

func ArrayInit[T any](a *Array[T], count int) {
	if count < 0 {
		count = 0
	}
	cap := 8
	if cap < count {
		cap = count
	}
	*a = Array[T]{
		data:     make([]T, cap),
		count:    count,
		capacity: cap,
	}
}

func ArrayInitWithCount[T any](a *Array[T], count int) {
	ArrayInit(a, count)
}

func ArrayInitWithCountAndCapacity[T any](a *Array[T], count, capacity int) {
	if count < 0 {
		count = 0
	}
	if capacity < count {
		capacity = count
	}
	*a = Array[T]{
		data:     make([]T, capacity),
		count:    count,
		capacity: capacity,
	}
}

func ArrayMake[T any]() Array[T] {
	var a Array[T]
	ArrayInit(&a, 0)
	return a
}

func ArrayMakeWithCount[T any](count int) Array[T] {
	var a Array[T]
	ArrayInit(&a, count)
	return a
}

func ArrayMakeWithCountAndCapacity[T any](count, capacity int) Array[T] {
	var a Array[T]
	ArrayInitWithCountAndCapacity(&a, count, capacity)
	return a
}

func ArrayMakeFromPtr[T any](data []T, count, capacity int) Array[T] {
	return Array[T]{data: data, count: count, capacity: capacity}
}

func ArrayFree[T any](a *Array[T]) {
	a.data = nil
	a.count = 0
	a.capacity = 0
}

func (a *Array[T]) grow(minCapacity int) {
	newCap := a.capacity
	if newCap < 8 {
		newCap = 8
	}
	for newCap < minCapacity {
		newCap *= 2
	}
	newData := make([]T, newCap)
	copy(newData, a.data[:a.count])
	a.data = newData
	a.capacity = newCap
}

func (a *Array[T]) Add(v T) {
	if a.count >= a.capacity {
		a.grow(a.count + 1)
	}
	a.data[a.count] = v
	a.count++
}

func (a *Array[T]) AddAndGet() *T {
	a.Add(*new(T))
	return &a.data[a.count-1]
}

func (a *Array[T]) AddElems(elems []T) {
	elemCount := len(elems)
	if elemCount == 0 {
		return
	}
	if a.capacity < a.count+elemCount {
		a.grow(a.count + elemCount)
	}
	copy(a.data[a.count:], elems)
	a.count += elemCount
}

func (a *Array[T]) Pop() T {
	if a.count <= 0 {
		panic("pop from empty array")
	}
	a.count--
	return a.data[a.count]
}

func (a *Array[T]) Clear() {
	a.count = 0
}

func (a *Array[T]) Reserve(capacity int) {
	if a.capacity >= capacity {
		return
	}
	newData := make([]T, capacity)
	copy(newData, a.data[:a.count])
	a.data = newData
	a.capacity = capacity
}

func (a *Array[T]) Resize(count int) {
	if a.capacity < count {
		a.grow(count)
	}
	a.count = count
}

func (a *Array[T]) SetCapacity(capacity int) {
	if a.capacity == capacity {
		return
	}
	if capacity < a.count {
		a.count = capacity
	}
	newData := make([]T, capacity)
	copy(newData, a.data[:a.count])
	a.data = newData
	a.capacity = capacity
}

func (a *Array[T]) Get(index int) T {
	if index < 0 || index >= a.count {
		panic("array index out of bounds")
	}
	return a.data[index]
}

func (a *Array[T]) Ptr(index int) *T {
	if index < 0 || index >= a.count {
		panic("array index out of bounds")
	}
	return &a.data[index]
}

func (a *Array[T]) Set(index int, v T) {
	if index < 0 || index >= a.count {
		panic("array index out of bounds")
	}
	a.data[index] = v
}

func (a *Array[T]) EndPtr() *T {
	if a.count > 0 {
		return &a.data[a.count-1]
	}
	return nil
}

func (a *Array[T]) Slice(lo, hi int) Slice[T] {
	if lo < 0 {
		lo = 0
	}
	if hi > a.count {
		hi = a.count
	}
	if lo > hi {
		lo = hi
	}
	return Slice[T]{
		data:  a.data[lo:hi],
		count: hi - lo,
	}
}

func (a *Array[T]) CloneInto(b *Array[T]) {
	b.data = make([]T, a.count)
	copy(b.data, a.data[:a.count])
	b.count = a.count
	b.capacity = a.count
}

func (a *Array[T]) Clone() Array[T] {
	var b Array[T]
	a.CloneInto(&b)
	return b
}

func (a *Array[T]) OrderedRemove(index int) {
	if index < 0 || index >= a.count {
		panic("array index out of bounds")
	}
	copy(a.data[index:], a.data[index+1:a.count])
	a.count--
}

func (a *Array[T]) UnorderedRemove(index int) {
	if index < 0 || index >= a.count {
		panic("array index out of bounds")
	}
	a.count--
	if index != a.count {
		a.data[index] = a.data[a.count]
	}
}

func (a *Array[T]) Copy(data Slice[T], offset int) {
	toCopy := data.count
	if toCopy > a.count-offset {
		toCopy = a.count - offset
	}
	if toCopy > 0 {
		copy(a.data[offset:], data.data[:toCopy])
	}
}

func (a *Array[T]) Sort(cmp func(T, T) int) {
	if a.count <= 1 {
		return
	}
	// Insertion sort for small arrays (good enough for this context).
	d := a.data[:a.count]
	for i := 1; i < len(d); i++ {
		key := d[i]
		j := i - 1
		for j >= 0 && cmp(d[j], key) > 0 {
			d[j+1] = d[j]
			j--
		}
		d[j+1] = key
	}
}

// --- Slice functions ---

func SliceFromArray[T any](a Array[T]) Slice[T] {
	return Slice[T]{
		data:  a.data,
		count: a.count,
	}
}

func SliceMake[T any](count int) Slice[T] {
	if count < 0 {
		count = 0
	}
	return Slice[T]{
		data:  make([]T, count),
		count: count,
	}
}

func SliceInit[T any](s *Slice[T], count int) {
	*s = SliceMake[T](count)
}

func SliceFree[T any](s *Slice[T]) {
	s.data = nil
	s.count = 0
}

func (s *Slice[T]) Resize(newCount int) {
	if s.count == newCount {
		return
	}
	if newCount == 0 {
		s.data = nil
		s.count = 0
		return
	}
	newData := make([]T, newCount)
	if s.count < newCount {
		copy(newData, s.data[:s.count])
	} else {
		copy(newData, s.data[:newCount])
	}
	s.data = newData
	s.count = newCount
}

func (s *Slice[T]) Get(index int) T {
	if index < 0 || index >= s.count {
		panic("slice index out of bounds")
	}
	return s.data[index]
}

func (s *Slice[T]) Set(index int, v T) {
	if index < 0 || index >= s.count {
		panic("slice index out of bounds")
	}
	s.data[index] = v
}

func (s *Slice[T]) Slice(lo, hi int) Slice[T] {
	if lo < 0 {
		lo = 0
	}
	if hi > s.count {
		hi = s.count
	}
	if lo > hi {
		lo = hi
	}
	return Slice[T]{
		data:  s.data[lo:hi],
		count: hi - lo,
	}
}

func (s *Slice[T]) Clone() Slice[T] {
	clone := make([]T, s.count)
	copy(clone, s.data[:s.count])
	return Slice[T]{
		data:  clone,
		count: s.count,
	}
}

func (s *Slice[T]) Copy(data Slice[T]) {
	n := s.count
	if n > data.count {
		n = data.count
	}
	copy(s.data[:n], data.data[:n])
}

func (s *Slice[T]) CopyOffset(data Slice[T], offset int) {
	if offset >= s.count {
		return
	}
	n := data.count
	if n > s.count-offset {
		n = s.count - offset
	}
	copy(s.data[offset:offset+n], data.data[:n])
}

func (s *Slice[T]) OrderedRemove(index int) {
	if index < 0 || index >= s.count {
		panic("slice index out of bounds")
	}
	copy(s.data[index:], s.data[index+1:s.count])
	s.count--
}

func (s *Slice[T]) UnorderedRemove(index int) {
	if index < 0 || index >= s.count {
		panic("slice index out of bounds")
	}
	s.count--
	if index != s.count {
		s.data[index] = s.data[s.count]
	}
}

// --- Iterator helpers ---

// Begin returns a pointer to the first element.
func ArrayBegin[T any](a *Array[T]) *T {
	if a.count == 0 {
		return nil
	}
	return &a.data[0]
}

func ArrayBeginConst[T any](a *Array[T]) *T {
	return ArrayBegin(a)
}

// End returns a pointer past the last element.
func ArrayEnd[T any](a *Array[T]) *T {
	if a.count == 0 {
		return nil
	}
	return &a.data[a.count]
}

func ArrayEndConst[T any](a *Array[T]) *T {
	return ArrayEnd(a)
}

func SliceBegin[T any](s *Slice[T]) *T {
	if s.count == 0 {
		return nil
	}
	return &s.data[0]
}

func SliceBeginConst[T any](s *Slice[T]) *T {
	return SliceBegin(s)
}

func SliceEnd[T any](s *Slice[T]) *T {
	if s.count == 0 {
		return nil
	}
	return &s.data[s.count]
}

func SliceEndConst[T any](s *Slice[T]) *T {
	return SliceEnd(s)
}
