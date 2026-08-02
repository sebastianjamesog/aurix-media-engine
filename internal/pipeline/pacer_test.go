package pipeline

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestFramePacer(t *testing.T) {
	pacer := NewFramePacer(10)

	var sentCount int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pacer.Start(ctx, func(frame []byte) error {
		atomic.AddInt32(&sentCount, 1)
		return nil
	})

	// Push 3 frames
	pacer.Push([]byte("frame1"))
	pacer.Push([]byte("frame2"))
	pacer.Push([]byte("frame3"))

	// Wait 70ms (should process 3 frames at 20ms intervals)
	time.Sleep(70 * time.Millisecond)
	pacer.Stop()

	count := atomic.LoadInt32(&sentCount)
	if count < 3 {
		t.Errorf("expected at least 3 frames sent, got %d", count)
	}
}
