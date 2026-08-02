package queue

import (
	"math/rand"
	"sync"
	"time"

	"aurix/internal/providers"
)

// RepeatMode represents the queue looping mode.
type RepeatMode int

const (
	RepeatOff RepeatMode = iota
	RepeatTrack
	RepeatQueue
)

// Queue maintains an ordered list of tracks to play for a single guild.
type Queue struct {
	tracks     []*providers.TrackInfo
	current    *providers.TrackInfo
	repeatMode RepeatMode
	rng        *rand.Rand
	mu         sync.RWMutex
}

// NewQueue creates a new track queue instance.
func NewQueue() *Queue {
	return &Queue{
		tracks: make([]*providers.TrackInfo, 0),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Add appends one or more tracks to the end of the queue.
func (q *Queue) Add(tracks ...*providers.TrackInfo) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, tracks...)
}

// Next pops and returns the next track to play based on current repeat mode.
func (q *Queue) Next() *providers.TrackInfo {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.repeatMode == RepeatTrack && q.current != nil {
		return q.current
	}

	if len(q.tracks) == 0 {
		q.current = nil
		return nil
	}

	nextTrack := q.tracks[0]
	q.tracks = q.tracks[1:]

	if q.repeatMode == RepeatQueue && q.current != nil {
		q.tracks = append(q.tracks, q.current)
	}

	q.current = nextTrack
	return nextTrack
}

// Current returns the currently playing track.
func (q *Queue) Current() *providers.TrackInfo {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.current
}

// Shuffle randomly reshuffles the upcoming tracks in the queue.
func (q *Queue) Shuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tracks) <= 1 {
		return
	}

	q.rng.Shuffle(len(q.tracks), func(i, j int) {
		q.tracks[i], q.tracks[j] = q.tracks[j], q.tracks[i]
	})
}

// SetRepeatMode updates the repeat mode (Off, Track, Queue).
func (q *Queue) SetRepeatMode(mode RepeatMode) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.repeatMode = mode
}

// Clear removes all upcoming tracks from the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = make([]*providers.TrackInfo, 0)
}

// List returns a snapshot slice of tracks in the queue.
func (q *Queue) List() []*providers.TrackInfo {
	q.mu.RLock()
	defer q.mu.RUnlock()
	snapshot := make([]*providers.TrackInfo, len(q.tracks))
	copy(snapshot, q.tracks)
	return snapshot
}

// Len returns the number of items in the queue.
func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tracks)
}
