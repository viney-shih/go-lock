package lock

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

// ================
// Benchmark
// ================

// goos: darwin
// goarch: amd64
// pkg: github.com/viney-shih/go-lock
//
// BenchmarkRWMutexLock-8                          42930610                29.0 ns/op             0 B/op          0 allocs/op
// BenchmarkConcurrentRWMutexLock-8                12205321               103 ns/op               0 B/op          0 allocs/op
// BenchmarkConcurrent50RWMutexLock-8               8393517               144 ns/op               0 B/op          0 allocs/op
//
// BenchmarkChanMutexLock-8                        21671048                54.7 ns/op             0 B/op          0 allocs/op
// BenchmarkConcurrentChanMutexLock-8               5118392               228 ns/op               0 B/op          0 allocs/op
// BenchmarkChanMutexTryLock-8                     21809853                55.1 ns/op             0 B/op          0 allocs/op
// BenchmarkConcurrentChanMutexTryLock-8           429987619                2.98 ns/op            0 B/op          0 allocs/op
//
// BenchmarkCASMutexLock-8                          7097668               158 ns/op              96 B/op          1 allocs/op
// BenchmarkConcurrentCASMutexLock-8                1793245               649 ns/op             255 B/op          3 allocs/op
// BenchmarkConcurrent50CASMutexLock-8              1727955               688 ns/op             255 B/op          3 allocs/op
// BenchmarkCASMutexTryLock-8                       7509704               156 ns/op              96 B/op          1 allocs/op
// BenchmarkConcurrentCASMutexTryLock-8            11179422               106 ns/op               4 B/op          0 allocs/op
// BenchmarkConcurrent50CASMutexTryLock-8           6191577               183 ns/op              37 B/op          0 allocs/op

func BenchmarkRWMutexLock(b *testing.B) {
	rwMut := sync.RWMutex{}

	for i := 0; i < b.N; i++ {
		rwMut.Lock()
		rwMut.Unlock()
	}
}

func BenchmarkConcurrentRWMutexLock(b *testing.B) {
	rwMut := sync.RWMutex{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rwMut.Lock()
			rwMut.Unlock()
		}
	})
}

func BenchmarkConcurrent50RWMutexLock(b *testing.B) {
	rwMut := sync.RWMutex{}
	rand.Seed(time.Now().UnixNano())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if rand.Intn(100) < 50 {
				rwMut.RLock()
				rwMut.RUnlock()
				continue
			}

			rwMut.Lock()
			rwMut.Unlock()
		}
	})
}

func BenchmarkChanMutexLock(b *testing.B) {
	chanMut := NewChanMutex()

	for i := 0; i < b.N; i++ {
		chanMut.Lock()
		chanMut.Unlock()
	}
}

func BenchmarkConcurrentChanMutexLock(b *testing.B) {
	chanMut := NewChanMutex()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			chanMut.Lock()
			chanMut.Unlock()
		}
	})
}

func BenchmarkChanMutexTryLock(b *testing.B) {
	chanMut := NewChanMutex()

	for i := 0; i < b.N; i++ {
		if chanMut.TryLock() {
			chanMut.Unlock()
		}
	}
}

func BenchmarkConcurrentChanMutexTryLock(b *testing.B) {
	chanMut := NewChanMutex()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if chanMut.TryLock() {
				chanMut.Unlock()
			}
		}
	})
}

func BenchmarkCASMutexLock(b *testing.B) {
	casMut := NewCASMutex()

	for i := 0; i < b.N; i++ {
		casMut.Lock()
		casMut.Unlock()
	}
}

func BenchmarkConcurrentCASMutexLock(b *testing.B) {
	casMut := NewCASMutex()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			casMut.Lock()
			casMut.Unlock()
		}
	})
}

func BenchmarkConcurrent50CASMutexLock(b *testing.B) {
	casMut := NewCASMutex()
	rand.Seed(time.Now().UnixNano())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if rand.Intn(100) < 50 {
				casMut.RLock()
				casMut.RUnlock()
				continue
			}

			casMut.Lock()
			casMut.Unlock()
		}
	})
}

func BenchmarkCASMutexTryLock(b *testing.B) {
	casMut := NewCASMutex()

	for i := 0; i < b.N; i++ {
		if casMut.TryLock() {
			casMut.Unlock()
		}
	}
}

func BenchmarkConcurrentCASMutexTryLock(b *testing.B) {
	casMut := NewCASMutex()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if casMut.TryLock() {
				casMut.Unlock()
			}
		}
	})
}

func BenchmarkConcurrent50CASMutexTryLock(b *testing.B) {
	casMut := NewCASMutex()
	rand.Seed(time.Now().UnixNano())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if rand.Intn(100) < 50 {
				if casMut.RTryLock() {
					casMut.RUnlock()
				}
				continue
			}

			if casMut.TryLock() {
				casMut.Unlock()
			}
		}
	})
}
