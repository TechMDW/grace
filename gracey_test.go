package grace

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"
)

func TestAddAndDone(t *testing.T) {
	Reset()
	Add()
	if Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", Count())
	}
	Done()
	if Count() != 0 {
		t.Errorf("Expected count to be 0, got %d", Count())
	}
}

func TestAddAndDoneGoroutine(t *testing.T) {
	Reset()
	go func() {
		for i := 0; i < 80000; i++ {
			Add()
		}
	}()
	time.Sleep(1 * time.Second)
	go func() {
		for i := 0; i < 80000; i++ {
			Done()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := WaitCtx(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if Count() != 0 {
		t.Errorf("Expected count to be 0, got %d", Count())
	}
}

func TestBlockAndUnblockAdd(t *testing.T) {
	Reset()
	BlockAdd()
	Add()
	Add()
	if Count() != 0 {
		t.Errorf("BlockAdd failed, count is %d", Count())
	}
	UnblockAdd()
	Add()
	if Count() != 1 {
		t.Errorf("UnblockAdd failed, count is %d", Count())
	}
}

func TestWaitFunction(t *testing.T) {
	Reset()
	Add()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		time.Sleep(5 * time.Second)
		Done()
	}()

	err := WaitCtx(ctx)
	if err != nil {
		if err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got %v", err)
		}
	}
	if Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", Count())
	}

	Reset()
	Add()

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		time.Sleep(5 * time.Millisecond)
		Done()
	}()

	err = WaitCtx(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if Count() != 0 {
		t.Errorf("Expected count to be 0, got %d", Count())
	}
}

func TestCountFunction(t *testing.T) {
	Reset()
	for i := 0; i < 10; i++ {
		Add()
	}
	if Count() != 10 {
		t.Errorf("Count is %d, expected 10", Count())
	}
	for i := 0; i < 5; i++ {
		Done()
	}
	if Count() != 5 {
		t.Errorf("Count is %d, expected 5", Count())
	}
	for i := 0; i < 5; i++ {
		Add()
	}
	if Count() != 10 {
		t.Errorf("Count is %d, expected 10", Count())
	}
	for i := 0; i < 10; i++ {
		Done()
	}
	if Count() != 0 {
		t.Errorf("Count is %d, expected 0", Count())
	}
}

func TestAddXAndDoneX(t *testing.T) {
	Reset()
	AddX(5)
	if Count() != 5 {
		t.Errorf("Expected count to be 5, got %d", Count())
	}
	DoneX(3)
	if Count() != 2 {
		t.Errorf("Expected count to be 2, got %d", Count())
	}
	DoneX(2)
	if Count() != 0 {
		t.Errorf("Expected count to be 0, got %d", Count())
	}

	Reset()
	AddX(math.MaxInt32)
	if Count() != int64(math.MaxInt32) {
		t.Errorf("Expected count to be %d, got %d", math.MaxInt32, Count())
	}
	AddX(1)
	if Count() != int64(math.MaxInt32)+1 {
		t.Errorf("Expected count to be %d, got %d", int64(math.MaxInt32)+1, Count())
	}

	Reset()
	AddX(math.MaxInt64)
	if Count() != math.MaxInt64 {
		t.Errorf("Expected count to be MaxInt64 (%d), got %d", math.MaxInt64, Count())
	}
	AddX(5000)
	if Count() != math.MaxInt64 {
		t.Errorf("Expected count to be %d, got %d", math.MaxInt64, Count())
	}
	DoneX(50000000)
	if count != math.MaxInt64-50000000 {
		t.Errorf("Expected count to be %d, got %d", math.MaxInt64-50000000, Count())
	}

	Reset()
	AddX(10)
	DoneX(15)
	if Count() != 0 {
		t.Errorf("Expected count to be 0 after underflow protection, got %d", Count())
	}
}

func TestGraceReset(t *testing.T) {
	Reset()

	AddX(5)

	Reset()

	var wg sync.WaitGroup

	expectedCountAfterReset := 3
	for i := 0; i < expectedCountAfterReset; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Add()
		}()
	}

	wg.Wait()

	actualCount := Count()
	if actualCount != int64(expectedCountAfterReset) {
		t.Errorf("expected count to be %d after reset, but got %d", expectedCountAfterReset, actualCount)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for i := 0; i < expectedCountAfterReset; i++ {
		Done()
	}

	if err := WaitCtx(ctx); err != nil {
		t.Errorf("WaitCtx returned an error: %v", err)
	}

	if finalCount := Count(); finalCount != 0 {
		t.Errorf("expected final count to be 0, but got %d", finalCount)
	}
}

func BenchmarkAddDone(b *testing.B) {
	Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Add()
		Done()
	}
}

func BenchmarkAddDoneX(b *testing.B) {
	Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddX(5)
		DoneX(5)
	}
}

func BenchmarkAddDoneConcurrent(b *testing.B) {
	Reset()
	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Add()
			Done()
		}()
	}

	wg.Wait()
}

func BenchmarkAddDoneXConcurrent(b *testing.B) {
	Reset()
	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			AddX(5)
			DoneX(5)
		}()
	}

	wg.Wait()
}
