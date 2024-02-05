package grace

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	forceShutdownTries = 2
	blockAdd           int32

	mu    sync.RWMutex
	count int64
	reset int64
)

func init() {
	count = 0
	reset = 0
}

// Adds 1 resource to wait for.
func Add() {
	AddX(1)
}

// Adds x resources to wait for.
func AddX(x int) {
	v := atomic.LoadInt64(&reset)
	mu.Lock()
	defer mu.Unlock()
	if v != reset {
		return
	}

	blocked := atomic.LoadInt32(&blockAdd)
	if blocked == 1 {
		return
	}

	if int64(x) > math.MaxInt64-count {
		count = math.MaxInt64
		return
	}

	count += int64(x)
}

// Mark 1 resource as done.
func Done() {
	DoneX(1)
}

// Mark x resources as done.
func DoneX(x int) {
	v := atomic.LoadInt64(&reset)
	mu.Lock()
	defer mu.Unlock()
	if v != reset {
		return
	}
	if int64(x) > count {
		count = 0
		return
	}

	count -= int64(x)
}

// Count returns the current amount of resources being waited for.
func Count() int64 {
	mu.RLock()
	defer mu.RUnlock()

	return count
}

// Reset resets Grace to its initial state.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	count = 0
	atomic.AddInt64(&reset, 1)
}

// IsEmpty checks if the count variable is 0.
func IsEmpty() bool {
	mu.RLock()
	defer mu.RUnlock()

	return count == 0
}

// BlockAdd preventing Add() operations from modifying the count.
func BlockAdd() {
	atomic.StoreInt32(&blockAdd, 1)
}

// UnblockAdd allows Add() operations to modify the count.
func UnblockAdd() {
	atomic.StoreInt32(&blockAdd, 0)
}

// Wait blocks until the count becomes 0, indicating all Add() and Del() operations are completed.
func Wait() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		if IsEmpty() {
			return
		}
	}
}

// WaitCtx blocks until the count becomes 0 or the provided context is canceled.
func WaitCtx(ctx context.Context) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if IsEmpty() {
				return nil
			}
		}
	}

	return nil
}

// SetForceShutdownTries sets the number of attempts to force shutdown after a SIGINT or SIGTERM is received.
//
//	This function only affects the grace.Grace() function.
func SetForceShutdownTries(tries int) {
	atomic.StoreInt32(&blockAdd, int32(tries))
}

// Grace blocks function execution until a SIGINT or SIGTERM is received. It will then wait for all resources to finish before returning.
// Pass a context to Grace to allow for a timeout. If the context is canceled, Grace will return immediately without waiting.
//
// If a SIGINT or SIGTERM is received again, Grace will force shutdown after forceShutdownTries attempts (default: 2).
//
// This function serves as more of a convenience/example function for a common use case. If you need more control over the shutdown process, you should use the Wait() or WaitCtx() functions.
func Grace(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigChan)
		close(sigChan)
	}()

	sig := <-sigChan
	go sigCalledAgain(sigChan, cancel)
	BlockAdd()
	fmt.Printf("Grace is now waiting after receiving: %v\n", sig)
	fmt.Printf("Grace is waiting for %d resources to finish\n", Count())

	go func(ctx context.Context) {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			select {
			case <-ctx.Done():
				fmt.Println("Grace is finished waiting")
				return
			default:
				fmt.Printf("Grace is waiting for %d resources to finish\n", Count())
			}
		}
	}(ctx)

	WaitCtx(ctx)
	os.Exit(0)
}

// Helper function to handle SIGINT and SIGTERM signals.
// If the signal is received again, it will force shutdown after forceShutdownTries attempts.
func sigCalledAgain(sigChan chan os.Signal, cancel context.CancelFunc) {
	var c int

	for {
		<-sigChan
		c++
		fmt.Printf("Grace force shutdown started %d/%d\n", c, forceShutdownTries)
		if c >= forceShutdownTries {
			fmt.Println("Grace is force shutting down")
			cancel()
			return
		}
	}
}
