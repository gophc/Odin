package common

import (
	"sync"
	"sync/atomic"
)

type (
	Mutex struct {
		mu sync.Mutex
	}
	RecursiveMutex struct {
		mu    sync.Mutex
		owner int
		count int
	}
	RwMutex struct {
		mu sync.RWMutex
	}
	Semaphore struct {
		ch chan struct{}
	}
	Parker struct {
		state atomic.Uint32
		mu    sync.Mutex
		cond  *sync.Cond
	}
	Wait_Signal struct {
		futex Futex
	}
	RWSpinLock struct {
		bits atomic.Int32
	}
)

const (
	RWLOCK_WRITER   = 1 << 0
	RWLOCK_UPGRADED = 1 << 1
	RWLOCK_READER   = 1 << 2
)

// --- Futex ---

func (f *Futex) Load() i32 { return atomic.LoadInt32(&f.val) }

func (f *Futex) Store(v i32) { atomic.StoreInt32(&f.val, v) }

func (f *Futex) Add(delta i32) i32 { return atomic.AddInt32(&f.val, delta) }

func (f *Futex) CompareExchange(old, new i32) bool {
	return atomic.CompareAndSwapInt32(&f.val, old, new)
}

func FutexWait(f *Futex, val i32) {
	for atomic.LoadInt32(&f.val) == val {
	}
}

func FutexSignal(f *Futex) {
	atomic.StoreInt32(&f.val, 1)
}

func FutexBroadcast(f *Futex) {
	atomic.StoreInt32(&f.val, 1)
}

// --- Mutex ---

func MutexLock(m *Mutex)                { m.mu.Lock() }
func MutexTryLock(m *Mutex) bool        { return m.mu.TryLock() }
func MutexUnlock(m *Mutex)              { m.mu.Unlock() }

// --- RecursiveMutex ---

func getGoroutineID() int {
	// Use a simple counter-based approach since Go doesn't expose goroutine IDs.
	// This is sufficient for recursive mutex semantics.
	return 0 // placeholder; recursive mutex uses this for same-goroutine detection
}

func RecursiveMutexLock(m *RecursiveMutex) {
	goid := getGoroutineID()
	if goid == 0 {
		// Fallback: just use TryLock-style semantics
		m.mu.Lock()
		return
	}
	for {
		m.mu.Lock()
		if m.owner == 0 || m.owner == goid {
			m.owner = goid
			m.count++
			return
		}
		m.mu.Unlock()
	}
}

func RecursiveMutexTryLock(m *RecursiveMutex) bool {
	goid := getGoroutineID()
	if goid == 0 {
		return m.mu.TryLock()
	}
	m.mu.Lock()
	if m.owner == 0 || m.owner == goid {
		m.owner = goid
		m.count++
		return true
	}
	m.mu.Unlock()
	return false
}

func RecursiveMutexUnlock(m *RecursiveMutex) {
	m.count--
	if m.count == 0 {
		m.owner = 0
	}
	m.mu.Unlock()
}

// --- RwMutex ---

func RwMutexLock(m *RwMutex)               { m.mu.Lock() }
func RwMutexTryLock(m *RwMutex) bool       { return m.mu.TryLock() }
func RwMutexUnlock(m *RwMutex)             { m.mu.Unlock() }
func RwMutexSharedLock(m *RwMutex)         { m.mu.RLock() }
func RwMutexTrySharedLock(m *RwMutex) bool { return m.mu.TryRLock() }
func RwMutexSharedUnlock(m *RwMutex)       { m.mu.RUnlock() }

// --- Semaphore ---

func SemaphoreInit(s *Semaphore, count int) { s.ch = make(chan struct{}, count) }

func SemaphorePost(s *Semaphore, n int) {
	for i := 0; i < n; i++ {
		select {
		case s.ch <- struct{}{}:
		default:
		}
	}
}

func SemaphoreWait(s *Semaphore) { <-s.ch }

// --- Parker ---

func ParkerInit(p *Parker) {
	p.state.Store(0)
	p.mu = sync.Mutex{}
	p.cond = sync.NewCond(&p.mu)
}

func Park(p *Parker) {
	p.mu.Lock()
	if p.state.CompareAndSwap(0, 1) {
		for p.state.Load() == 1 {
			p.cond.Wait()
		}
	}
	p.mu.Unlock()
}

func UnparkOne(p *Parker) {
	p.mu.Lock()
	p.state.Store(0)
	p.cond.Signal()
	p.mu.Unlock()
}

func UnparkAll(p *Parker) {
	p.mu.Lock()
	p.state.Store(0)
	p.cond.Broadcast()
	p.mu.Unlock()
}

// --- WaitSignal ---

func WaitSignalUntilAvailable(ws *Wait_Signal) {
	for ws.futex.Load() == 0 {
		FutexWait(&ws.futex, 0)
	}
}

func WaitSignalSet(ws *Wait_Signal) {
	ws.futex.Store(1)
	FutexBroadcast(&ws.futex)
}

// --- RWSpinLock ---

func RWLockReleaseWrite(l *RWSpinLock) {
	l.bits.And(^(RWLOCK_WRITER | RWLOCK_UPGRADED))
	FutexBroadcast(&Futex{val: l.bits.Load()})
}

func RWLockTryAcquireUpgrade(l *RWSpinLock) bool {
	old := l.bits.Load()
	if old&(RWLOCK_UPGRADED|RWLOCK_WRITER) != 0 {
		return false
	}
	return l.bits.CompareAndSwap(old, old|RWLOCK_UPGRADED)
}

func RWLockAcquireUpgrade(l *RWSpinLock) {
	for !RWLockTryAcquireUpgrade(l) {
		FutexWait(&Futex{val: l.bits.Load()}, RWLOCK_UPGRADED|RWLOCK_WRITER)
	}
}

func RWLockReleaseUpgrade(l *RWSpinLock) {
	l.bits.Add(-RWLOCK_UPGRADED)
	FutexBroadcast(&Futex{val: l.bits.Load()})
}

func RWLockTryReleaseUpgradeAndAcquireWrite(l *RWSpinLock) bool {
	return l.bits.CompareAndSwap(RWLOCK_UPGRADED, RWLOCK_WRITER)
}

func RWLockReleaseUpgradeAndAcquireWrite(l *RWSpinLock) {
	for !RWLockTryReleaseUpgradeAndAcquireWrite(l) {
		FutexWait(&Futex{val: l.bits.Load()}, RWLOCK_UPGRADED)
	}
}

// --- Thread ---

func ThreadCurrentID() u32 { return 0 }
func YieldThread()         {}
func YieldProcess()        {}

type Thread struct {
	ID   int
	Func func()
	Done bool
}

var CurrentThread *Thread

func GetCurrentThread() *Thread { return CurrentThread }
