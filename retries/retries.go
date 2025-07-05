package retries

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"math"
	"math/rand"
	"time"
)

type AllAt interface{}

type RetryableFuncObject struct {
	err    error
	ctx    context.Context
	logger *zap.SugaredLogger
	rest   []AllAt
}

type RetryableFunc func() RetryableFuncObject

const (
	maxElapsedTime       = 30 * time.Second      //Maximum total duration that all retries may last
	initialSleepInterval = 50 * time.Millisecond //The initial waiting time before the first retry
	maxSleepInterval     = 5 * time.Second       //The longest possible waiting time between two retries
	maxRetries           = 10                    //Upper limit for the number of retries if time is not the primary termination condition
)

var ErrMaxElapsedTimeExceeded = errors.New("max elapsed time for operation exceeded")
var ErrRetriesExhausted = errors.New("operation failed after all retries exhausted")
var ErrTransient = errors.New("transient error, retry")

// The RetryWithExponentialBackoff simply repeats the function calls
// and passes the error or success it receives from the RetryableFunc at the end
func RetryWithExponentialBackoff(rf RetryableFunc) RetryableFuncObject {
	startTime := time.Now()
	sleep := initialSleepInterval
	var rfo RetryableFuncObject
	for i := 0; i <= maxRetries; i++ {
		rfo = rf()
		if rfo.err == ErrTransient {
			// Check if there are still attempts available for Retry
			if i == maxRetries {
				rfo.logger.Errorf("Operation failed after all retries exhausted")
				rfo.err = ErrRetriesExhausted
				break
			}
			// Error can possibly be fixed by repeating
			// sleep = min(cap, random_between(base, sleep * 3)
			sleep = time.Duration(math.Min(float64(maxSleepInterval), float64(initialSleepInterval)+rand.Float64()*float64(3*sleep-initialSleepInterval)))
			rfo.logger.Infof("Waiting for %v before next retry attempt for (attempt %d)", sleep, i+1)
			//Proactive check whether the maxSleepInterval would be exceeded after waiting sleep-long
			if time.Since(startTime)+sleep >= maxElapsedTime {
				rfo.logger.Errorf("Operation timed out after %v", maxElapsedTime)
				rfo.err = ErrMaxElapsedTimeExceeded
				break
			}
			select {
			case <-time.After(sleep):
			case <-rfo.ctx.Done():
				rfo.logger.Errorf("Operation cancelled during backoff: %v", rfo.ctx.Err())
				rfo.err = rfo.ctx.Err()
				break
			}
		} else {
			// rfo.err == nil or rfo.err is a permanent error
			break
		}
	}
	return rfo
}
