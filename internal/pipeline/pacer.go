package pipeline

import (
	"context"
	"sync"
	"time"
)

const (
	FrameDuration = 20 * time.Millisecond
	SamplesPer20ms = 960 // At 48kHz
)

// FramePacer delivers Opus frame packets at strict 20ms intervals.
type FramePacer struct {
	frameChan chan []byte
	stopChan  chan struct{}
	mu        sync.Mutex
	running   bool
}

// NewFramePacer initializes a frame pacer with a ring buffer channel.
func NewFramePacer(bufferSize int) *FramePacer {
	if bufferSize <= 0 {
		bufferSize = 25 // 25 * 20ms = 500ms pre-buffer
	}
	return &FramePacer{
		frameChan: make(chan []byte, bufferSize),
		stopChan:  make(chan struct{}),
	}
}

// Push adds an Opus audio frame packet to the pacer buffer.
func (p *FramePacer) Push(frame []byte) bool {
	select {
	case p.frameChan <- frame:
		return true
	default:
		// Ring buffer full; drop frame or handle buffer overflow
		return false
	}
}

// Start begins the high-precision 20ms frame delivery ticker loop.
func (p *FramePacer) Start(ctx context.Context, sendFunc func(frame []byte) error) {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.mu.Unlock()

	ticker := time.NewTicker(FrameDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-ticker.C:
			select {
			case frame := <-p.frameChan:
				if sendFunc != nil {
					_ = sendFunc(frame)
				}
			default:
				// Silence frame / buffer underrun
			}
		}
	}
}

// Stop terminates the frame pacer ticker loop.
func (p *FramePacer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		p.running = false
		close(p.stopChan)
	}
}
