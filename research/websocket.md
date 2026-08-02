# WebSocket Protocol Specification

## Overview

The WebSocket protocol in Aurix Media Engine (AME) provides full-duplex real-time events between the AME audio node and client applications (Discord bots or Cluster Gateways).

---

## 1. Connection Lifecycle

### URL & Headers
* **URL**: `ws://<host>:<port>/v4/websocket`
* **Headers**:
  ```http
  Authorization: <server_password>
  User-Id: <bot_user_id>
  Client-Name: <bot_name/version>
  Session-Id: <optional_existing_session_id_to_resume>
  ```

---

## 2. Server -> Client Events (`op: "event"`)

Every event message emitted over WebSocket follows this structure:

```json
{
  "op": "event",
  "type": "<EventType>",
  "guildId": "123456789012345678"
}
```

### 1. `TrackStartEvent`
Emitted when a track starts playing.
```json
{
  "op": "event",
  "type": "TrackStartEvent",
  "guildId": "123456789012345678",
  "track": {
    "encoded": "QAAAjQIAAAAAAAAAC2V4YW1wbGUudXBzA...",
    "info": { "title": "Track Title", "author": "Author", "length": 180000 }
  }
}
```

### 2. `TrackEndEvent`
Emitted when a track finishes playing or is stopped/replaced.
```json
{
  "op": "event",
  "type": "TrackEndEvent",
  "guildId": "123456789012345678",
  "track": { "encoded": "..." },
  "reason": "finished" // "finished" | "loadFailed" | "stopped" | "replaced" | "cleanup"
}
```

### 3. `TrackExceptionEvent`
Emitted when an error occurs during track decoding or streaming.
```json
{
  "op": "event",
  "type": "TrackExceptionEvent",
  "guildId": "123456789012345678",
  "track": { "encoded": "..." },
  "exception": {
    "message": "Failed to decode WebM stream from provider",
    "severity": "common", // "common" | "suspicious" | "fault"
    "cause": "HTTP 403 Forbidden"
  }
}
```

### 4. `TrackStuckEvent`
Emitted if a track takes longer than the threshold (e.g. 10000ms) to deliver audio frames.
```json
{
  "op": "event",
  "type": "TrackStuckEvent",
  "guildId": "123456789012345678",
  "track": { "encoded": "..." },
  "thresholdMs": 10000
}
```

### 5. `WebSocketClosedEvent`
Emitted when Discord closes the underlying voice connection.
```json
{
  "op": "event",
  "type": "WebSocketClosedEvent",
  "guildId": "123456789012345678",
  "code": 4014,
  "reason": "Voice channel deleted",
  "byRemote": true
}
```

---

## 3. Server -> Client Telemetry (`op: "playerUpdate"` & `op: "stats"`)

### Player Update Payload
Sent periodically (every 5 seconds) per active playing guild:
```json
{
  "op": "playerUpdate",
  "guildId": "123456789012345678",
  "state": {
    "time": 1690000000000,
    "position": 45200,
    "connected": true,
    "ping": 18
  }
}
```

### Engine Stats Payload
Sent periodically to connected clients:
```json
{
  "op": "stats",
  "players": 42,
  "playingPlayers": 30,
  "uptime": 1204000,
  "memory": {
    "free": 150000000,
    "used": 35000000,
    "allocated": 185000000,
    "reservable": 512000000
  },
  "cpu": {
    "cores": 4,
    "systemLoad": 0.08,
    "lavalinkLoad": 0.02
  },
  "frameStats": {
    "sent": 15000,
    "nulled": 0,
    "deficit": 0
  }
}
```

---

## 4. Session Ready & Resuming

When a client connects to the WebSocket:
1. If no `Session-Id` header is provided, the server generates a new random `sessionId` (e.g. `c9f8a...`) and sends `ready`:
```json
{
  "op": "ready",
  "resumed": false,
  "sessionId": "c9f8a1b2c3d4e5f6"
}
```

2. If a client disconnects unexpectedly, the server keeps players alive for the session resume window. When the client reconnects passing `Session-Id: c9f8a...`:
```json
{
  "op": "ready",
  "resumed": true,
  "sessionId": "c9f8a1b2c3d4e5f6"
}
```
All event subscriptions and players are restored immediately without audio interruption.
