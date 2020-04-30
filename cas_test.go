package lock

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type casSuite struct {
	suite.Suite
}

func (s *casSuite) SetupSuite() {}

func (s *casSuite) TearDownSuite() {}

func (s *casSuite) SetupTest() {}

func (s *casSuite) TearDownTest() {}

func TestCASSuite(t *testing.T) {
	suite.Run(t, new(casSuite))
}

func (s *casSuite) TestCASTryLock() {
	casMut := NewCASMutex()

	// write lock then write lock
	s.Require().True(casMut.TryLock())
	s.Require().False(casMut.TryLock())
	casMut.Unlock()

	// write lock then read lock
	s.Require().True(casMut.TryLock())
	s.Require().False(casMut.RTryLock())
	casMut.Unlock()

	// read lock then write lock
	s.Require().True(casMut.RTryLock())
	s.Require().False(casMut.TryLock())
	casMut.RUnlock()

	// read lock then read lock
	s.Require().True(casMut.RTryLock())
	s.Require().True(casMut.RTryLock())
	casMut.RUnlock()
	casMut.RUnlock()
}

func (s *casSuite) TestCASTryLockWithTimeout() {
	casMut := NewCASMutex()

	// write lock then write lock
	s.Require().True(casMut.TryLockWithTimeout(50 * time.Millisecond))
	s.Require().False(casMut.TryLockWithTimeout(50 * time.Millisecond))
	casMut.Unlock()

	// write lock then read lock
	s.Require().True(casMut.TryLockWithTimeout(50 * time.Millisecond))
	s.Require().False(casMut.RTryLockWithTimeout(50 * time.Millisecond))
	casMut.Unlock()

	// read lock then write lock
	s.Require().True(casMut.RTryLockWithTimeout(50 * time.Millisecond))
	s.Require().False(casMut.TryLockWithTimeout(50 * time.Millisecond))
	casMut.RUnlock()

	// read lock then read lock
	s.Require().True(casMut.RTryLockWithTimeout(50 * time.Millisecond))
	s.Require().True(casMut.RTryLockWithTimeout(50 * time.Millisecond))
	casMut.RUnlock()
	casMut.RUnlock()
}

func (s *casSuite) TestInvalidCASUnlock() {
	casMut := NewCASMutex()
	defer func() {
		r := recover()
		s.Require().NotNil(r)
		s.Require().Equal("Unlock failed", r)
	}()

	casMut.RLock()
	casMut.Unlock()
}

func (s *casSuite) TestInvalidCASRUnlock() {
	casMut := NewCASMutex()
	defer func() {
		r := recover()
		s.Require().NotNil(r)
		s.Require().Equal("RUnlock failed", r)
	}()

	casMut.Lock()
	casMut.RUnlock()
}

func (s *casSuite) TestCASLockRacing() {
	casMut := NewCASMutex()
	count := int32(0) // default value

	// write lock then write lock
	casMut.Lock()

	go func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&count, 1) // A = 1
		casMut.Unlock()
	}()

	casMut.Lock()
	s.Require().Equal(int32(2), atomic.AddInt32(&count, 1)) // add 1 after A
	casMut.Unlock()

	// write lock then read lock
	casMut.Lock()
	go func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&count, 1) // B = 3
		casMut.Unlock()
	}()

	casMut.RLock()
	s.Require().Equal(int32(4), atomic.AddInt32(&count, 1)) // add 1 after B
	casMut.RUnlock()

	// read lock then write lock
	casMut.RLock()
	go func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&count, 1) // C = 5
		casMut.RUnlock()
	}()

	casMut.Lock()
	s.Require().Equal(int32(6), atomic.AddInt32(&count, 1)) // add 1 after C
	casMut.Unlock()

	// read lock then read lock
	casMut.RLock()
	go func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&count, 1) // add 1 after D
		casMut.RUnlock()
	}()

	casMut.RLock()
	s.Require().Equal(int32(7), atomic.AddInt32(&count, 1)) // D
	casMut.RUnlock()
}

func (s *casSuite) TestCASNoStarve() {
	casMut := NewCASMutex()
	pause := make(chan bool)
	result := make(chan int32, 2)
	count := int32(0) // default value

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		casMut.RLock()
		atomic.AddInt32(&count, 1) // A = 1
		<-pause
		casMut.RUnlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(10 * time.Millisecond)
		casMut.RLock()
		atomic.AddInt32(&count, 2) // B = 3
		<-pause
		casMut.RUnlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(50 * time.Millisecond)
		casMut.Lock()                        // blocked until A and B are done
		result <- atomic.AddInt32(&count, 3) // C = 6
		casMut.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Millisecond)
		casMut.RLock()                       // blocked until C is done
		result <- atomic.AddInt32(&count, 4) // D = 10
		casMut.RUnlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(150 * time.Millisecond)
		s.Require().False(casMut.TryLock())
		s.Require().False(casMut.TryLockWithTimeout(20 * time.Millisecond))
		s.Require().False(casMut.RTryLock())
		s.Require().False(casMut.RTryLockWithTimeout(20 * time.Millisecond))
	}()

	time.Sleep(300 * time.Millisecond)
	close(pause)
	s.Require().Equal(int32(6), <-result)
	s.Require().Equal(int32(10), <-result)

	wg.Wait()
}

func (s *casSuite) TestCASRLockRacing() {
	casMut := NewCASMutex()
	wg := &sync.WaitGroup{}
	numOfTasks := 10
	numOfTimes := 1000

	for i := 0; i < numOfTasks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for inner := 0; inner < numOfTimes; inner++ {
				casMut.RLock()
				time.Sleep(10 * time.Millisecond)
				casMut.RUnlock()
			}

		}()
	}

	wg.Wait()
}
