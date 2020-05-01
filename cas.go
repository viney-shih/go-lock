package lock

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sync/semaphore"
)

// CASMutex provides interfaces of read-write spinlock and read-write trylock with CAS mechanism.
type CASMutex interface {
	// Lock acquires the write lock.
	// If it is currently held by others, Lock will wait until it has a chance to acquire it.
	Lock()
	// TryLock attempts to acquire the write lock without blocking.
	// Return false if someone is holding it now.
	TryLock() bool
	// TryLockWithTimeout attempts to acquire the write lock within a period of time.
	// Return false if spending time is more than duration and no chance to acquire it.
	TryLockWithTimeout(time.Duration) bool
	// Unlock releases the write lock
	Unlock()

	// RLock acquires the read lock.
	// If it is currently held by others writing, RLock will wait until it has a chance to acquire it.
	RLock()
	// RTryLock attempts to acquire the read lock without blocking.
	// Return false if someone is writing it now.
	RTryLock() bool
	// RTryLockWithTimeout attempts to acquire the read lock within a period of time.
	// Return false if spending time is more than duration and no chance to acquire it.
	RTryLockWithTimeout(time.Duration) bool
	// RUnlock releases the read lock
	RUnlock()
}

type casState int32

const (
	casStateUndefined casState = iota - 2 // -2
	casStateWriteLock                     // -1
	casStateNoLock                        // 0
	casStateReadLock                      // >= 1
)

func (m *casMutex) getState(n int32) casState {
	switch st := casState(n); {
	case st == casStateWriteLock:
		fallthrough
	case st == casStateNoLock:
		return st
	case st >= casStateReadLock:
		return casStateReadLock
	default:
		// actually, it should not happened.
		return casStateUndefined
	}
}

type casMutex struct {
	state     casState
	turnstile *semaphore.Weighted

	broadcastChan chan struct{}
	broadcastMut  sync.RWMutex
}

func (m *casMutex) listen() <-chan struct{} {
	m.broadcastMut.RLock()
	defer m.broadcastMut.RUnlock()

	return m.broadcastChan
}

func (m *casMutex) broadcast() {
	newCh := make(chan struct{})

	m.broadcastMut.Lock()
	ch := m.broadcastChan
	m.broadcastChan = newCh
	m.broadcastMut.Unlock()

	close(ch)
}

func (m *casMutex) Lock() {
	ctx := context.Background()
	m.turnstile.Acquire(ctx, 1)
	defer m.turnstile.Release(1)

	m.tryLock(ctx)
}

func (m *casMutex) tryLock(ctx context.Context) bool {
	for {
		broker := m.listen()
		if atomic.CompareAndSwapInt32(
			(*int32)(unsafe.Pointer(&m.state)),
			int32(casStateNoLock),
			int32(casStateWriteLock),
		) {
			return true
		}

		if ctx == nil {
			return false
		}

		select {
		case <-ctx.Done():
			// timeout
			return false
		case <-broker:
			// waiting for signal triggered by m.broadcast() and trying again.
		}
	}
}

func (m *casMutex) TryLock() bool {
	if !m.turnstile.TryAcquire(1) {
		return false
	}

	defer m.turnstile.Release(1)

	return m.tryLock(nil)
}

func (m *casMutex) TryLockWithTimeout(duration time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	if err := m.turnstile.Acquire(ctx, 1); err != nil {
		// Acquire failed due to timeout
		return false
	}

	defer m.turnstile.Release(1)

	return m.tryLock(ctx)
}

func (m *casMutex) Unlock() {
	if ok := atomic.CompareAndSwapInt32(
		(*int32)(unsafe.Pointer(&m.state)),
		int32(casStateWriteLock),
		int32(casStateNoLock),
	); !ok {
		panic("Unlock failed")
	}

	m.broadcast()
}

func (m *casMutex) RLock() {
	ctx := context.Background()
	m.turnstile.Acquire(ctx, 1)
	m.turnstile.Release(1)

	m.rTryLock(ctx)
}

func (m *casMutex) rTryLock(ctx context.Context) bool {
	for {
		broker := m.listen()
		n := atomic.LoadInt32((*int32)(unsafe.Pointer(&m.state)))
		st := m.getState(n)
		switch st {
		case casStateNoLock, casStateReadLock:
			if atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.state)), n, n+1) {
				return true
			}
		}

		if ctx == nil {
			return false
		}

		select {
		case <-ctx.Done():
			// timeout
			return false
		default:
			switch st {
			// read-lock failed due to concurrence issue, try again immediately
			case casStateNoLock, casStateReadLock:
				runtime.Gosched() // allow other goroutines to do stuff.
				continue
			}
		}

		select {
		case <-ctx.Done():
			// timeout
			return false
		case <-broker:
			// waiting for signal triggered by m.broadcast() and trying again.
		}
	}
}

func (m *casMutex) RTryLock() bool {
	if !m.turnstile.TryAcquire(1) {
		return false
	}

	m.turnstile.Release(1)

	return m.rTryLock(nil)
}

func (m *casMutex) RTryLockWithTimeout(duration time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	if err := m.turnstile.Acquire(ctx, 1); err != nil {
		// Acquire failed due to timeout
		return false
	}

	m.turnstile.Release(1)

	return m.rTryLock(ctx)
}

func (m *casMutex) RUnlock() {
	n := atomic.AddInt32((*int32)(unsafe.Pointer(&m.state)), -1)
	switch m.getState(n) {
	case casStateUndefined, casStateWriteLock:
		panic("RUnlock failed")
	case casStateNoLock:
		m.broadcast()
	}
}

// NewCASMutex returns CASMutex lock
func NewCASMutex() CASMutex {
	return &casMutex{
		state:         casStateNoLock,
		turnstile:     semaphore.NewWeighted(1),
		broadcastChan: make(chan struct{}),
	}
}
