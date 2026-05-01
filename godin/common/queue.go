package common

// MPSCNode is a node in a Multi-Producer Single-Consumer queue.
type MPSCNode[T any] struct {
	next  *MPSCNode[T]
	value T
}

// MPSCQueue is a Multi-Producer Single-Consumer queue.
type MPSCQueue[T any] struct {
	sentinel *MPSCNode[T]
	head     **MPSCNode[T]
	tail     *MPSCNode[T]
	count    int
}

func MPSCInit[T any](q *MPSCQueue[T]) {
	sentinel := &MPSCNode[T]{}
	q.sentinel = sentinel
	q.head = &sentinel.next
	q.tail = sentinel
	q.count = 0
}

func MPSCDestroy[T any](q *MPSCQueue[T]) {
	q.sentinel = nil
	q.head = nil
	q.tail = nil
	q.count = 0
}

func MPSCAllocNode[T any](q *MPSCQueue[T], value T) *MPSCNode[T] {
	return &MPSCNode[T]{value: value}
}

func MPSCFreeNode[T any](q *MPSCQueue[T], node *MPSCNode[T]) {
	// GC handles this in Go.
}

func MPSCEnqueue[T any](q *MPSCQueue[T], node *MPSCNode[T]) {
	node.next = nil
	// Non-blocking: append to tail.
	tail := q.tail
	tail.next = node
	q.tail = node
	q.count++
}

func MPSCEnqueueValue[T any](q *MPSCQueue[T], value T) {
	node := MPSCAllocNode(q, value)
	MPSCEnqueue(q, node)
}

func MPSCDequeue[T any](q *MPSCQueue[T]) (T, bool) {
	head := *q.head
	if head == nil {
		var zero T
		return zero, false
	}
	*q.head = head.next
	if *q.head == nil {
		q.tail = q.sentinel
	}
	q.count--
	return head.value, true
}

// --- MPMCQueue: Multi-Producer Multi-Consumer bounded queue ---

type MPMCQueue[T any] struct {
	nodes    []T
	indices  []i32
	capacity i32
	mask     i32
	head     Futex
	tail     Futex
}

func MPMCInit[T any](q *MPMCQueue[T], size int) {
	if size < 8 {
		size = 8
	}
	size = nextPow2Int(size)
	mask := i32(size - 1)
	q.nodes = make([]T, size)
	q.indices = make([]i32, size)
	q.capacity = i32(size)
	q.mask = mask
	q.head = Futex{val: 0}
	q.tail = Futex{val: 0}
	for i := i32(0); i < i32(size); i++ {
		q.indices[i] = i
	}
}

func MPMCDestroy[T any](q *MPMCQueue[T]) {
	q.nodes = nil
	q.indices = nil
}

func MPMCEnqueue[T any](q *MPMCQueue[T], value T) bool {
	for {
		head := q.head.Load()
		index := head & q.mask
		nodeIdx := q.indices[index]
		diff := nodeIdx - head
		if diff == 0 {
			if q.head.CompareExchange(head, head+1) {
				q.nodes[index] = value
				q.indices[index] = head + 1
				return true
			}
		} else if diff < 0 {
			if q.capacity-q.count() <= 0 {
				return false
			}
		}
	}
}

func MPMCDequeue[T any](q *MPMCQueue[T]) (T, bool) {
	for {
		tail := q.tail.Load()
		if tail >= q.head.Load() {
			var zero T
			return zero, false
		}
		index := tail & q.mask
		nodeIdx := q.indices[index]
		diff := nodeIdx - (tail + 1)
		if diff == 0 {
			if q.tail.CompareExchange(tail, tail+1) {
				val := q.nodes[index]
				q.indices[index] = tail + q.mask + 1
				return val, true
			}
		} else if diff < 0 {
			return MPMCDequeue(q)
		}
	}
}

func (q *MPMCQueue[T]) count() i32 {
	return q.head.Load() - q.tail.Load()
}
