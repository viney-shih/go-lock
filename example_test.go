package lock_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/viney-shih/go-lock"
)

func ExampleCASMutex() {
	// set RWMutex with CAS mechanism (CASMutex).
	var rwMut lock.RWMutex = lock.NewCASMutex()
	// set default value
	count := int32(0)

	// block here
	rwMut.Lock()
	go func() {
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Now is", atomic.AddInt32(&count, 1)) // Now is 1
		rwMut.Unlock()
	}()

	// waiting for previous goroutine releasing the lock, and locking it again
	rwMut.Lock()
	fmt.Println("Now is", atomic.AddInt32(&count, 2)) // Now is 3

	// TryLock without blocking
	// Return false, because the lock is not released.
	fmt.Println("Return", rwMut.TryLock())

	// RTryLockWithTimeout without blocking
	// Return false, because the lock is not released.
	fmt.Println("Return", rwMut.RTryLockWithTimeout(50*time.Millisecond))

	// TryLockWithContext without blocking
	ctx, cancel := context.WithTimeout(context.TODO(), 50*time.Millisecond)
	defer cancel()
	// Return false, because the lock is not released.
	fmt.Println("Return", rwMut.TryLockWithContext(ctx))

	// release the lock in the end.
	rwMut.Unlock()

	// Output:
	// Now is 1
	// Now is 3
	// Return false
	// Return false
	// Return false
}
