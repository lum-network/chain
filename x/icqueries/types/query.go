package types

import (
	"time"
)

// HasTimedOut Check if a query has timed-out by checking whether the block time is after  the timeout timestamp
func (q *Query) HasTimedOut(currentBlockTime time.Time) bool {
	return q.TimeoutTimestamp < uint64(currentBlockTime.UnixNano())
}
