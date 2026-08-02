package voice

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync/atomic"
)

// UDPConn manages the UDP voice socket connection to Discord.
type UDPConn struct {
	conn       *net.UDPConn
	ssrc       uint32
	seq        uint32
	timestamp  uint32
	secretKey  [32]byte
	publicIP   string
	publicPort uint16
}

// NewUDPConn connects to Discord's Voice UDP endpoint.
func NewUDPConn(address string, ssrc uint32) (*UDPConn, error) {
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP socket: %w", err)
	}

	return &UDPConn{
		conn: conn,
		ssrc: ssrc,
	}, nil
}

// SetSecretKey updates the 32-byte encryption key received from Session Description.
func (u *UDPConn) SetSecretKey(key [32]byte) {
	u.secretKey = key
}

// DiscoverIP performs the 74-byte STUN IP discovery packet handshake with Discord.
func (u *UDPConn) DiscoverIP() (string, uint16, error) {
	// Construct 74-byte IP Discovery Packet
	packet := make([]byte, 74)
	binary.BigEndian.PutUint16(packet[0:2], 1)  // Type: Request (1)
	binary.BigEndian.PutUint16(packet[2:4], 70) // Length: 70
	binary.BigEndian.PutUint32(packet[4:8], u.ssrc)

	if _, err := u.conn.Write(packet); err != nil {
		return "", 0, fmt.Errorf("failed to send IP discovery packet: %w", err)
	}

	response := make([]byte, 74)
	n, err := u.conn.Read(response)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read IP discovery response: %w", err)
	}

	if n < 74 {
		return "", 0, fmt.Errorf("invalid IP discovery response length: %d", n)
	}

	// Extract Null-terminated IP string from bytes 8 to 71
	ipBytes := response[8:68]
	end := 0
	for ; end < len(ipBytes); end++ {
		if ipBytes[end] == 0 {
			break
		}
	}

	u.publicIP = string(ipBytes[:end])
	u.publicPort = binary.BigEndian.Uint16(response[72:74])

	return u.publicIP, u.publicPort, nil
}

// CreateRTPHeader constructs a standard 12-byte RTP header.
func (u *UDPConn) CreateRTPHeader() []byte {
	header := make([]byte, 12)
	header[0] = 0x80 // Version 2
	header[1] = 0x78 // Payload type 120 (Opus)

	seq := atomic.AddUint32(&u.seq, 1)
	binary.BigEndian.PutUint16(header[2:4], uint16(seq))

	ts := atomic.AddUint32(&u.timestamp, 960) // 20ms frame = 960 samples
	binary.BigEndian.PutUint32(header[4:8], ts)

	binary.BigEndian.PutUint32(header[8:12], u.ssrc)

	return header
}

// Close closes the underlying UDP connection.
func (u *UDPConn) Close() error {
	if u.conn != nil {
		return u.conn.Close()
	}
	return nil
}
