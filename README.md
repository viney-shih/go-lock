# go-lock
**go-lock** is a Golang library implementing an effcient read-write lock with the following built-in mechanism:
- Spinlock
- Trylock
- No-starve read-write solution

Native`sync/Mutex` and `sync/RWMutex` are very powerful and reliable spinlock in Golang. However, it became a disaster if the lock was not released as expected or someone was holding the lock too long at the peak time. Dealing with those cases, **go-lock** provides `TryLock` and `TryLockWithTimeout` function in addition to `Lock` and `Unlock`.

## Installation

```sh
    go get github.com/viney-shih/go-lock
```

## Example
```go
package main

import (
    "fmt"

    lock "github.com/viney-shih/go-lock"
)

func main() {
    // init with default value
    casMut := lock.NewCASMutex()
    count := int32(0)

    casMut.Lock()
    go func() {
        time.Sleep(50 * time.Millisecond)
        fmt.Println("Now is", atomic.AddInt32(&count, 1)) // Now is 1
        casMut.Unlock()
    }()

    casMut.Lock()
    fmt.Println("Now is", atomic.AddInt32(&count, 2)) // Now is 3
    fmt.Println("Return", casMut.TryLock()) // Return false without blocking
    fmt.Println("Return", casMut.RTryLockWithTimeout(50 * time.Millisecond)) // Return false without blocking
    casMut.Unlock()

    // Output:
    // Now is 1
    // Now is 3
    // Return false
    // Return false
}
```

more [examples](./cas_test.go)

## License
[Apache-2.0](https://opensource.org/licenses/Apache-2.0)
