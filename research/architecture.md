# System Architecture Specification

## Overview

The **Aurix Media Engine (AME)** is designed as a modular, concurrency-first, high-throughput distributed media engine. It acts as an audio streaming middleware between Discord Bot applications (or external clients) and Discord's Voice Infrastructure.

---

## 1. High-Level Architecture Diagram

```
                              ┌────────────────────────┐
                              │  Discord Bot Client(s) │
                              └───────────┬────────────┘
                                          │ REST / WebSocket
                                          ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                             AURIX MEDIA ENGINE (AME)                             │
│                                                                                  │
│   ┌──────────────────────────────────────────────────────────────────────────┐   │
│   │                          API & Gateway Module                            │   │
│   │   - REST Router (Fiber / Gin / net/http)                                 │   │
│   │   - WebSocket Handler & Session Manager                                  │   │
│   └────────────────────────────────────┬─────────────────────────────────────┘   │
│                                        │                                         │
│                                        ▼                                         │
│   ┌──────────────────────────────────────────────────────────────────────────┐   │
│   │                          Player Manager Module                           │   │
│   │   - Active Guild Player Registry                                         │   │
│   │   - State Transitions (Play, Pause, Stop, Seek, Destroy)                 │   │
│   │   - Queue & Scheduler Coordinator                                        │   │
│   └───────────────┬────────────────────┬────────────────────┬────────────────┘   │
│                   │                    │                    │                    │
│                   ▼                    ▼                    ▼                    │
│    ┌────────────────────┐   ┌────────────────────┐   ┌────────────────────┐      │
│    │ Provider Registry  │   │  Audio Pipeline    │   │ Voice Manager      │      │
│    │ - Youtube          │   │ - Audio Decoder    │   │ - Voice WS Gateway │      │
│    │ - Spotify          │   │ - DSP Filter Chain │   │ - UDP RTP Sender   │      │
│    │ - SoundCloud       │   │ - Opus Encoder     │   │ - Encryption       │      │
│    │ - HTTP / Local     │   │ - 20ms Frame Pacer │   │ - IP Discovery     │      │
│    └────────────────────┘   └────────────────────┘   └────────────────────┘      │
│                                                                                  │
└────────────────────────────────────────┬─────────────────────────────────────────┘
                                         │ UDP Encrypted RTP Packets (Opus)
                                         ▼
                              ┌────────────────────────┐
                              │  Discord Voice Server  │
                              └────────────────────────┘
```

---

## 2. Core Internal Subsystems

### 1. Gateway & Session Manager (`internal/gateway`)
* **REST Server**: Handles HTTP endpoints for track resolution, player lifecycle, and stats.
* **WebSocket Server**: Manages full-duplex client connections, event broadcasting (`TrackStart`, `TrackEnd`), and session state resumption.

### 2. Player Manager & Player State (`internal/player`)
* Manages a thread-safe `sync.Map` of active guild players.
* Coordinates playback controls (Play, Pause, Volume, Seek, Filters).
* Coordinates track queues and repeat/shuffle/autoplay modes.

### 3. Provider Manager (`internal/providers`)
* Isolated plugin architecture for resolving identifiers into playable streams.
* Unified `SourceProvider` interface:
  ```go
  type SourceProvider interface {
      Name() string
      CanHandle(identifier string) bool
      Resolve(ctx context.Context, identifier string) (*TrackResult, error)
      GetStream(ctx context.Context, track *Track) (io.ReadCloser, error)
  }
  ```

### 4. Audio Pipeline & DSP Engine (`internal/pipeline` & `internal/filters`)
* Ingests raw stream -> Decodes to 48kHz Stereo PCM S16LE.
* Applies dynamic DSP filter graphs (Equalizer, Pitch, Speed, Nightcore, Karaoke).
* Encodes PCM -> Opus frames (20ms length, 960 samples/channel).
* Feeds ring buffer connected to high-precision packet ticker.

### 5. Voice Connection Manager (`internal/voice`)
* Connects to Discord Voice WebSocket v8.
* Performs UDP STUN IP Discovery.
* Negotiates modern ciphers (`aead_xchacha20_poly1305_rtpsize`, `xsalsa20_poly1305`).
* Packetizes RTP + Opus and sends over UDP socket every 20ms.

---

## 3. Concurrency Model & Goroutine Boundaries

AME enforces clean concurrency isolation for every active player:

* **1 Player = 3 Isolated Goroutines**:
  1. **Voice WS Goroutine**: Listens for Discord Voice Gateway control events & heartbeats.
  2. **Decoder/Pipeline Goroutine**: Reads audio stream, applies DSP filters, encodes Opus, and fills RingBuffer.
  3. **UDP Sender Goroutine**: High-precision 20ms ticker reading from RingBuffer and sending UDP packets to Discord.

This goroutine isolation prevents slow network stream reads from causing packet drops on the UDP voice sender.

---

## 4. Clustering & Distribution Roadmap (`internal/cluster`)

For large-scale deployments (Phase 8), AME supports multi-node clustering:

```
              ┌────────────────────────┐
              │    Master Gateway      │ (Load Balancer / Ingress Router)
              └───────────┬────────────┘
                          │
            ┌─────────────┼─────────────┐
            ▼             ▼             ▼
      ┌──────────┐  ┌──────────┐  ┌──────────┐
      │ Node A   │  │ Node B   │  │ Node C   │
      └──────────┘  └──────────┘  └──────────┘
```

* **State Synchronization**: Shared Redis state / PubSub for active sessions and player allocation.
* **Auto Failover**: If Node A crashes, Master Gateway automatically migrates player sessions to Node B, resuming Discord voice handshakes without dropping bot connections.
