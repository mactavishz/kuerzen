package retries

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"math"
	"math/rand"
	"time"
)

// AllAt is a flexible type that acts as a container for diverse data types.
// This is used for holding various results of an operation when returned.
type AllAt interface{}

type RetryableFuncObject struct {
	Err    error
	Ctx    context.Context
	Logger *zap.SugaredLogger
	Rest   []AllAt //If more than the error must be returned, different return values can be contained here.
}

// RetryableFunc defines the signature for any operation that can be retried.
// This function type is designed to be implemented by a Closure.
// A Closure allows capturing necessary data.
// By using this standardized signature, several different functions (e.g. CreateShortURL, GetLongURL, SendURLCreationEvent)
// can be used within the RetryWithExponentialBackoff logic, regardless of their original parameters or their specific task.
type RetryableFunc func() RetryableFuncObject

const (
	maxElapsedTime       = 30 * time.Second      // Maximum total duration that all retries may last (including the initial attempt)
	initialSleepInterval = 50 * time.Millisecond // The initial waiting time before the first retry (base)
	maxSleepInterval     = 5 * time.Second       // The longest possible waiting time between two retries (cap)
	maxRetries           = 10                    // Upper limit for the number of retries if time is not the primary termination condition
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
	// Iteration i=0 is the initial attempt; iterations i=1 to maxRetries are subsequent retries.
	for i := 0; i <= maxRetries; i++ {
		rfo = rf()
		if rfo.Err == ErrTransient {
			// Error can possibly be fixed by repeating
			// Check if there are still attempts available for Retry
			if i == maxRetries {
				rfo.Logger.Errorf("Operation failed after all retries exhausted")
				rfo.Err = ErrRetriesExhausted
				break
			}
			// The first retry takes place after initialSleepInterval
			// A new sleep does not have to be calculated until the first retry
			if i != 0 {
				// sleep = min(cap, random_between(base, sleep * 3)
				sleep = time.Duration(math.Min(float64(maxSleepInterval), float64(initialSleepInterval)+rand.Float64()*float64(3*sleep-initialSleepInterval)))
			}
			rfo.Logger.Infof("Waiting for %v before next retry attempt for (attempt %d)", sleep, i+1)
			// Proactive check whether the maxSleepInterval would be exceeded after waiting sleep-long
			if time.Since(startTime)+sleep >= maxElapsedTime {
				rfo.Logger.Errorf("Operation timed out after %v", maxElapsedTime)
				rfo.Err = ErrMaxElapsedTimeExceeded
				break
			}
			select {
			// Successfully waited for the backoff duration. Proceed to the next iteration (retry)
			case <-time.After(sleep):
			// Listen for cancellation of the context used for the current operation attempt.
			case <-rfo.Ctx.Done():
				rfo.Logger.Errorf("Operation cancelled during backoff: %v", rfo.Ctx.Err())
				rfo.Err = rfo.Ctx.Err()
				break
			}
		} else {
			// rfo.err == nil or rfo.err is a permanent error
			break
		}
	}
	return rfo
}
