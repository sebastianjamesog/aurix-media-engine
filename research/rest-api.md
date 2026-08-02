# REST API Specification

## Overview

The Aurix Media Engine (AME) REST API allows Discord bots, cluster gateways, and dashboard management tools to control players, load tracks from audio providers, query engine stats, and manage sessions.

AME provides a **native v1 REST API** while remaining **compatible with Lavalink v4 REST endpoints** to allow existing bot libraries (e.g. LavaSrc, Lavalink.js, Disnake, Wavelink) to connect seamlessly.

---

## 1. Authentication & Base Headers

All requests require an authorization header:

```http
Authorization: <your_configured_server_password>
Content-Type: application/json
```

---

## 2. Track Loading & Resolution (`/v4/loadtracks`)

### Endpoint
`GET /v4/loadtracks?identifier=<query_or_url>`

### Query Parameters
* `identifier` (string, required): Search query or URL (e.g., `ytsearch:never gonna give you up`, `https://www.youtube.com/watch?v=...`, `https://open.spotify.com/track/...`, `https://example.com/audio.mp3`).

### Response Structure (`200 OK`)

```json
{
  "loadType": "track",
  "data": {
    "encoded": "QAAAjQIAAAAAAAAAC2V4YW1wbGUudXBzA...",
    "info": {
      "identifier": "dQw4w9WgXcQ",
      "isSeekable": true,
      "author": "Rick Astley",
      "length": 213000,
      "isStream": false,
      "position": 0,
      "title": "Never Gonna Give You Up",
      "uri": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
      "artworkUrl": "https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg",
      "isrc": "GBARL9700017",
      "sourceName": "youtube"
    },
    "pluginInfo": {}
  }
}
```

### Possible Load Types
* `track`: Single track loaded.
* `playlist`: Playlist loaded (contains `info` object and array of `tracks`).
* `search`: Search query results (array of `tracks`).
* `empty`: No matches found for the identifier.
* `error`: Resolution failed (contains `cause` and `message`).

---

## 3. Player Management (`/v4/sessions/{sessionId}/players`)

### 1. Get All Players
* **Endpoint**: `GET /v4/sessions/{sessionId}/players`
* **Response**: Array of active player state objects.

### 2. Get Player State
* **Endpoint**: `GET /v4/sessions/{sessionId}/players/{guildId}`

### 3. Update / Create Player
* **Endpoint**: `PATCH /v4/sessions/{sessionId}/players/{guildId}?noReplace=false`
* **Request Body**:
```json
{
  "track": {
    "encoded": "QAAAjQIAAAAAAAAAC2V4YW1wbGUudXBzA..."
  },
  "position": 0,
  "endTime": 0,
  "volume": 100,
  "paused": false,
  "filters": {
    "volume": 1.0,
    "equalizer": [
      { "band": 0, "gain": 0.2 },
      { "band": 1, "gain": 0.15 }
    ],
    "karaoke": {
      "level": 1.0,
      "monoLevel": 1.0,
      "filterBand": 220.0,
      "filterWidth": 100.0
    },
    "timescale": {
      "speed": 1.0,
      "pitch": 1.2,
      "rate": 1.0
    }
  },
  "voice": {
    "token": "voice_token_from_discord",
    "endpoint": "voice_endpoint.discord.gg",
    "sessionId": "discord_voice_session_id"
  }
}
```

### 4. Destroy Player
* **Endpoint**: `DELETE /v4/sessions/{sessionId}/players/{guildId}`
* **Response**: `204 No Content`

---

## 4. Engine Statistics (`/v4/stats` / `/v1/stats`)

### Endpoint
`GET /v4/stats`

### Response
```json
{
  "players": 142,
  "playingPlayers": 98,
  "uptime": 86400000,
  "memory": {
    "free": 184549376,
    "used": 42107264,
    "allocated": 226656640,
    "reservable": 536870912
  },
  "cpu": {
    "cores": 8,
    "systemLoad": 0.12,
    "lavalinkLoad": 0.03
  },
  "frameStats": {
    "sent": 294020,
    "nulled": 12,
    "deficit": 2
  }
}
```

---

## 5. Session Operations (`/v4/sessions/{sessionId}`)

### Update Session (Resumability & Timeout)
* **Endpoint**: `PATCH /v4/sessions/{sessionId}`
* **Request Body**:
```json
{
  "resuming": true,
  "timeout": 60
}
```
* **Description**: Configures session resuming so that if a bot client disconnects, players remain playing audio for `timeout` seconds waiting for reconnection.
