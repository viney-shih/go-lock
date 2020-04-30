package lock

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type chanSuite struct {
	suite.Suite
}

func (s *chanSuite) SetupSuite() {}

func (s *chanSuite) TearDownSuite() {}

func (s *chanSuite) SetupTest() {}

func (s *chanSuite) TearDownTest() {}

func TestChanSuite(t *testing.T) {
	suite.Run(t, new(chanSuite))
}

func (s *chanSuite) TestChanTryLock() {
	chanMut := NewChanMutex()

	// write lock then write lock
	s.Require().True(chanMut.TryLock())
	s.Require().False(chanMut.TryLock())
	chanMut.Unlock()
}

func (s *chanSuite) TestChanTryLockWithTimeout() {
	chanMut := NewChanMutex()

	// write lock then write lock
	s.Require().True(chanMut.TryLockWithTimeout(50 * time.Millisecond))
	s.Require().False(chanMut.TryLockWithTimeout(50 * time.Millisecond))
	chanMut.Unlock()
}

func (s *chanSuite) TestChanLockRacing() {
	chanMut := NewChanMutex()
	count := int32(0) // default value

	// write lock then write lock
	chanMut.Lock()
	go func() {
		time.Sleep(50 * time.Millisecond)
		s.Require().Equal(int32(1), atomic.AddInt32(&count, 1)) // A
		chanMut.Unlock()
	}()

	chanMut.Lock()
	s.Require().Equal(int32(2), atomic.AddInt32(&count, 1)) // add 1 after A
	chanMut.Unlock()
}
