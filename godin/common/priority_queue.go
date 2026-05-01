package common

// PriorityQueue is a min-heap backed by an Array.
type PriorityQueue[T any] struct {
	queue Array[T]
	cmp   func(T, T) int
}

// PriorityQueueShiftDown restores heap property by sinking element at i0.
func PriorityQueueShiftDown[T any](pq *PriorityQueue[T], i0, n int) bool {
	if i0 < 0 || i0 >= n {
		return false
	}
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 {
			break
		}
		j := j1
		j2 := j1 + 1
		if j2 < n && pq.cmp(pq.queue.data[j2], pq.queue.data[j1]) < 0 {
			j = j2
		}
		if pq.cmp(pq.queue.data[j], pq.queue.data[i]) >= 0 {
			break
		}
		pq.queue.data[i], pq.queue.data[j] = pq.queue.data[j], pq.queue.data[i]
		i = j
	}
	return i > i0
}

// PriorityQueueShiftUp restores heap property by bubbling up element at j.
func PriorityQueueShiftUp[T any](pq *PriorityQueue[T], j int) {
	for j > 0 {
		i := (j - 1) / 2
		if i == j || pq.cmp(pq.queue.data[j], pq.queue.data[i]) >= 0 {
			break
		}
		pq.queue.data[i], pq.queue.data[j] = pq.queue.data[j], pq.queue.data[i]
		j = i
	}
}

// PriorityQueueFix restores heap property at index i.
func PriorityQueueFix[T any](pq *PriorityQueue[T], i int) {
	if !PriorityQueueShiftDown(pq, i, pq.queue.count) {
		PriorityQueueShiftUp(pq, i)
	}
}

// PriorityQueuePush adds a value to the heap.
func PriorityQueuePush[T any](pq *PriorityQueue[T], value T) {
	pq.queue.Add(value)
	PriorityQueueShiftUp(pq, pq.queue.count-1)
}

// PriorityQueuePop removes and returns the minimum element.
func PriorityQueuePop[T any](pq *PriorityQueue[T]) T {
	if pq.queue.count <= 0 {
		panic("pop from empty priority queue")
	}
	n := pq.queue.count - 1
	pq.queue.data[0], pq.queue.data[n] = pq.queue.data[n], pq.queue.data[0]
	PriorityQueueShiftDown(pq, 0, n)
	return pq.queue.Pop()
}

// PriorityQueueRemove removes and returns the element at index i.
func PriorityQueueRemove[T any](pq *PriorityQueue[T], i int) T {
	if i < 0 || i >= pq.queue.count {
		panic("index out of bounds")
	}
	n := pq.queue.count - 1
	if n != i {
		pq.queue.data[i], pq.queue.data[n] = pq.queue.data[n], pq.queue.data[i]
		PriorityQueueShiftDown(pq, i, n)
		PriorityQueueShiftUp(pq, i)
	}
	return pq.queue.Pop()
}

// PriorityQueueCreate builds a heap from an existing Array.
func PriorityQueueCreate[T any](queue Array[T], cmp func(T, T) int) PriorityQueue[T] {
	pq := PriorityQueue[T]{queue: queue, cmp: cmp}
	n := pq.queue.count
	for i := n/2 - 1; i >= 0; i-- {
		PriorityQueueShiftDown(&pq, i, n)
	}
	return pq
}
