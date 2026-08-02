# Discord Voice Protocol Specification

## Overview

The Discord Voice protocol allows bots and clients to establish encrypted real-time audio channels with Discord's Voice Servers. It operates across two communication channels:

1. **Voice Gateway (WebSocket)**: Controls session setup, heartbeats, IP discovery, encryption handshake, and voice state signaling.
2. **Voice UDP Socket**: Transmits raw audio frames wrapped in standard **RTP (Real-time Transport Protocol)** packets using symmetric encryption (e.g. `xsalsa20_poly1305` or `aead_xchacha20_poly1305_rtpsize`).

---

## 1. Connection Lifecycle

```
Client / AME                    Discord Gateway              Discord Voice Server
     │                                │                                │
     │── (1) Update Voice State ─────>│                                │
     │<── (2) VOICE_STATE_UPDATE ─────│                                │
     │<── (3) VOICE_SERVER_UPDATE ────│                                │
     │                                                                 │
     │── (4) Connect WebSocket (wss://<endpoint>?v=8) ────────────────>│
     │<── (5) Opcode 8 Hello (Heartbeat Interval) ────────────────────>│
     │── (6) Opcode 0 Identify (Server ID, User ID, Session, Token) ──>│
     │<── (7) Opcode 2 Ready (SSRC, IP, Port, Modes) ─────────────────>│
     │                                                                 │
     │── (8) Send UDP IP Discovery Packet ────────────────────────────>│
     │<── (9) Receive Public IP/Port Response ─────────────────────────│
     │                                                                 │
     │── (10) Opcode 1 Select Protocol (IP, Port, Encryption Mode) ───>│
     │<── (11) Opcode 4 Session Description (Secret Key) ──────────────│
     │                                                                 │
     │── (12) Opcode 5 Speaking (SSRC, Speaking Flags) ───────────────>│
     │── (13) Stream Encrypted Opus RTP Packets (UDP 20ms) ───────────>│
```

---

## 2. Voice Gateway (WebSocket) Details

* **Endpoint**: `wss://<endpoint_from_voice_server_update>?v=8`
* **Encoding**: JSON

### Key Opcodes

| Opcode | Name | Direction | Description |
| :--- | :--- | :--- | :--- |
| `0` | Identify | Client -> Server | Sends authentication token, session ID, user ID, server ID. |
| `1` | Select Protocol | Client -> Server | Selects UDP transport protocol and chosen encryption cipher. |
| `2` | Ready | Server -> Client | Returns SSRC (Synchronization Source ID), Voice IP, UDP Port, and supported encryption modes. |
| `3` | Heartbeat | Client -> Server | Keeps WebSocket connection alive based on interval. |
| `4` | Session Description | Server -> Client | Contains the 32-byte shared secret key used for UDP packet encryption. |
| `5` | Speaking | Client -> Server | Signals whether client is transmitting audio (Speaking flag `1 << 0`). |
| `8` | Hello | Server -> Client | Sent on initial connection; specifies `heartbeat_interval` in ms. |
| `13` | Hello / Ack | Server -> Client | Heartbeat acknowledgement. |

---

## 3. UDP Socket & IP Discovery

1. After receiving **Opcode 2 (Ready)**, open a UDP socket to the IP and Port provided in the payload.
2. Send a 74-byte **IP Discovery Packet** to perform STUN/NAT traversal:
   * First 2 bytes: Request type (`0x0001` - Request)
   * Bytes 2-4: Message length (`70` bytes)
   * Bytes 4-8: SSRC (Big-Endian uint32)
   * Remaining bytes: Zero-padded up to 74 bytes.
3. Read the 74-byte response packet from the UDP socket to extract your public IP address and external UDP port.

---

## 4. Encryption Handshake & Modes

Send **Opcode 1 (Select Protocol)**:
```json
{
  "op": 1,
  "d": {
    "protocol": "udp",
    "data": {
      "address": "<YOUR_PUBLIC_IP>",
      "port": 12345,
      "mode": "aead_xchacha20_poly1305_rtpsize"
    }
  }
}
```

### Supported Ciphers (Preferred Hierarchy)
1. `aead_xchacha20_poly1305_rtpsize` (Modern, hardware accelerated / secure)
2. `aead_aes256_gcm_rtpsize`
3. `xsalsa20_poly1305` (Legacy default)

After sending Opcode 1, the Voice Gateway returns **Opcode 4 (Session Description)** containing a `secret_key` (32 bytes).

---

## 5. Audio Framing & RTP Encapsulation

Audio must be encoded as **Opus (48kHz, Stereo, 20ms frame length)**.

### RTP Header (12 Bytes)

| Offset | Size | Field | Value / Description |
| :--- | :--- | :--- | :--- |
| `0` | 1 byte | Version & Flags | `0x80` (RTP v2) |
| `1` | 1 byte | Payload Type | `0x78` (120 - Opus) |
| `2-3` | 2 bytes | Sequence Number | Big-endian uint16 (increments by 1 per 20ms packet) |
| `4-7` | 4 bytes | Timestamp | Big-endian uint32 (increments by 960 per 20ms frame at 48kHz) |
| `8-11` | 4 bytes | SSRC | Big-endian uint32 (assigned in Opcode 2 Ready) |

### Encryption Execution

1. Prepare 12-byte RTP header + Opus Payload.
2. Generate Nonce according to selected cipher mode:
   * For `xsalsa20_poly1305`: 24-byte Nonce (12-byte RTP header + 12 zero-bytes).
   * For `aead_xchacha20_poly1305_rtpsize`: 24-byte Nonce (4-byte packet counter appended).
3. Encrypt the Opus frame payload using Secret Key + Nonce.
4. Append Nonce / Tag metadata as required by the chosen cipher mode.
5. Transmit the complete UDP packet every 20ms precision ticker.

---

## 6. AME Implementation Architecture (`internal/voice`)

```
               ┌────────────────────────┐
               │    Voice Connection    │
               └───────────┬────────────┘
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
    ┌─────────────────┐         ┌─────────────────┐
    │ Voice WebSocket │         │   Voice UDP     │
    │  - Heartbeat    │         │  - Packetizer   │
    │  - Signaling    │         │  - Encryptor    │
    │  - Reconnect    │         │  - 20ms Ticker  │
    └─────────────────┘         └─────────────────┘
```
