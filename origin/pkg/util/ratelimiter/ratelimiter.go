package ratelimiter

import (
	kcache "github.com/openshift/kubernetes/pkg/client/cache"
	"github.com/openshift/kubernetes/pkg/util/flowcontrol"
	utilruntime "github.com/openshift/kubernetes/pkg/util/runtime"
	utilwait "github.com/openshift/kubernetes/pkg/util/wait"
)

// HandlerFunc defines function signature for a RateLimitedFunction.
type HandlerFunc func() error

// RateLimitedFunction is a rate limited function controlling how often the function/handler is invoked.
type RateLimitedFunction struct {
	// Handler is the function to rate limit calls to.
	Handler HandlerFunc

	// Internal queue of requests to be processed.
	queue kcache.Queue

	// Rate limiting configuration.
	flowcontrol.RateLimiter
}

// NewRateLimitedFunction creates a new rate limited function.
func NewRateLimitedFunction(keyFunc kcache.KeyFunc, interval int, handlerFunc HandlerFunc) *RateLimitedFunction {
	fifo := kcache.NewFIFO(keyFunc)

	qps := float32(1000.0) // Call rate per second (SLA).
	if interval > 0 {
		qps = float32(1.0 / float32(interval))
	}

	limiter := flowcontrol.NewTokenBucketRateLimiter(qps, 1)

	return &RateLimitedFunction{handlerFunc, fifo, limiter}
}

// RunUntil begins processes the resources from queue asynchronously until
// stopCh is closed.
func (rlf *RateLimitedFunction) RunUntil(stopCh <-chan struct{}) {
	go utilwait.Until(rlf.pop, 0, stopCh)
}

// handleOne processes a request in the queue invoking the rate limited
// function.
func (rlf *RateLimitedFunction) pop() {
	rlf.RateLimiter.Accept()
	if _, err := rlf.queue.Pop(func(_ interface{}) error {
		return rlf.Handler()
	}); err != nil {
		utilruntime.HandleError(err)
	}
}

// Invoke adds a request if its not already present and waits for the
// background processor to execute it.
func (rlf *RateLimitedFunction) Invoke(resource interface{}) {
	rlf.queue.AddIfNotPresent(resource)
}