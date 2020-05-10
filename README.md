# go-lock

[![GoDoc](https://godoc.org/github.com/viney-shih/go-lock?status.svg)](https://godoc.org/github.com/viney-shih/go-lock)
[![GoDev](https://img.shields.io/badge/go.dev-doc-007d9c?style=flat-square&logo=read-the-docs)](https://pkg.go.dev/github.com/viney-shih/go-lock?tab=doc)
[![Build Status](https://travis-ci.com/viney-shih/go-lock.svg?branch=master)](https://travis-ci.com/github/viney-shih/go-lock)
[![Go Report Card](https://goreportcard.com/badge/github.com/viney-shih/go-lock)](https://goreportcard.com/report/github.com/viney-shih/go-lock)
[![Codecov](https://codecov.io/gh/viney-shih/go-lock/branch/master/graph/badge.svg)](https://codecov.io/gh/viney-shih/go-lock)
[![Coverage Status](https://coveralls.io/repos/github/viney-shih/go-lock/badge.svg?branch=master)](https://coveralls.io/github/viney-shih/go-lock?branch=master)
[![License](http://img.shields.io/badge/License-Apache_2-red.svg?style=flat)](http://www.apache.org/licenses/LICENSE-2.0)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#utilities)

**go-lock** is a Golang library implementing an effcient read-write lock with the following built-in mechanism:
- Mutex
- Trylock
- No-starve read-write solution

Native `sync/Mutex` and `sync/RWMutex` are very powerful and reliable. However, it became a disaster if the lock was not released as expected. Or someone was holding the lock too long at the peak time. It slowed down whole system. Dealing with those cases, **go-lock** implements TryLock, TryLockWithTimeout and TryLockWithContext function in addition to Lock and Unlock. It provides flexibility to control the resources.

## Installation

```sh
go get github.com/viney-shih/go-lock
```

## Example
```go
package main

import (
	"fmt"
	"sync/atomic"
	"time"

	lock "github.com/viney-shih/go-lock"
)

func main() {
	// initialized with default value
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

	// TryLock without blocking
	fmt.Println("Return", casMut.TryLock()) // Return false

	// RTryLockWithTimeout without blocking
	fmt.Println("Return", casMut.RTryLockWithTimeout(50*time.Millisecond)) // Return false
	casMut.Unlock()

	// Output:
	// Now is 1
	// Now is 3
	// Return false
	// Return false
}
```

- [More examples](./cas_test.go)
- [Full API documentation](https://godoc.org/github.com/viney-shih/go-lock)

## References
- https://github.com/golang/go/issues/6123
- https://github.com/LK4D4/trylock
- https://github.com/OneOfOne/go-utils/tree/master/sync
- https://github.com/lrita/gosync
- https://github.com/google/netstack/blob/master/tmutex/tmutex.go
- https://github.com/subchen/go-trylock

## License
[Apache-2.0](https://opensource.org/licenses/Apache-2.0)
