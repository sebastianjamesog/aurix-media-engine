package queue

import (
	"testing"

	"aurix/internal/providers"
)

func TestQueueOperations(t *testing.T) {
	q := NewQueue()

	track1 := &providers.TrackInfo{Title: "Track 1", Identifier: "t1"}
	track2 := &providers.TrackInfo{Title: "Track 2", Identifier: "t2"}

	q.Add(track1, track2)

	if q.Len() != 2 {
		t.Fatalf("expected queue len 2, got %d", q.Len())
	}

	next1 := q.Next()
	if next1.Identifier != "t1" {
		t.Errorf("expected t1, got %s", next1.Identifier)
	}

	if q.Current().Identifier != "t1" {
		t.Errorf("expected current t1, got %s", q.Current().Identifier)
	}

	next2 := q.Next()
	if next2.Identifier != "t2" {
		t.Errorf("expected t2, got %s", next2.Identifier)
	}

	next3 := q.Next()
	if next3 != nil {
		t.Errorf("expected nil for empty queue, got %v", next3)
	}
}

func TestQueueRepeatTrack(t *testing.T) {
	q := NewQueue()
	track1 := &providers.TrackInfo{Title: "Track 1", Identifier: "t1"}
	q.Add(track1)

	q.SetRepeatMode(RepeatTrack)

	first := q.Next()
	second := q.Next()

	if first.Identifier != "t1" || second.Identifier != "t1" {
		t.Errorf("expected repeated t1 track, got %s and %s", first.Identifier, second.Identifier)
	}
}
