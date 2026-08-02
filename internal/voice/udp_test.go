package voice

import (
	"encoding/binary"
	"testing"
)

func TestRTPHeaderCreation(t *testing.T) {
	conn := &UDPConn{
		ssrc: 0x12345678,
	}

	header := conn.CreateRTPHeader()

	if len(header) != 12 {
		t.Fatalf("expected 12 bytes RTP header, got %d", len(header))
	}

	if header[0] != 0x80 {
		t.Errorf("expected version 0x80, got 0x%x", header[0])
	}

	if header[1] != 0x78 {
		t.Errorf("expected payload type 0x78 (120), got 0x%x", header[1])
	}

	seq := binary.BigEndian.Uint16(header[2:4])
	if seq != 1 {
		t.Errorf("expected initial sequence 1, got %d", seq)
	}

	ts := binary.BigEndian.Uint32(header[4:8])
	if ts != 960 {
		t.Errorf("expected initial timestamp 960, got %d", ts)
	}

	ssrc := binary.BigEndian.Uint32(header[8:12])
	if ssrc != 0x12345678 {
		t.Errorf("expected ssrc 0x12345678, got 0x%x", ssrc)
	}
}
