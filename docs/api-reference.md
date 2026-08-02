# Aurix Media Engine — API & WebSocket Reference

## Base Server Configuration

* **Host**: `http://localhost:2333`
* **WebSocket Endpoint**: `ws://localhost:2333/v4/websocket` or `ws://localhost:2333/`
* **Default Authorization Password**: `youshallnotpass`

---

## 1. REST Endpoints

### Track Resolution (`GET /v4/loadtracks`)

```http
GET /v4/loadtracks?identifier=ytsearch:never+gonna+give+you+up
Authorization: youshallnotpass
```

#### Response Types
* `track`: Single track resolved.
* `playlist`: Playlist resolved.
* `search`: Search results returned.
* `empty`: No matches found.
* `error`: Resolution failed.

---

### Node Statistics (`GET /v4/stats`)

```http
GET /v4/stats
Authorization: youshallnotpass
```

Returns total active players, uptime, and memory usage telemetry.

---

### Get Active Players (`GET /v4/sessions/{sessionId}/players`)

```http
GET /v4/sessions/c9f8a1b2/players
Authorization: youshallnotpass
```

---

## 2. WebSocket Protocol Events

WebSocket clients receive real-time JSON payloads:

### `ready` Payload
```json
{
  "op": "ready",
  "resumed": false,
  "sessionId": "c9f8a1b2c3d4e5f6"
}
```

### `TrackStartEvent`
```json
{
  "op": "event",
  "type": "TrackStartEvent",
  "guildId": "123456789012345678",
  "track": { "info": { "title": "Track Title" } }
}
```
