<div align="center">

# 🎵 Aurix Media Engine (AME)

**The Next-Generation, Distributed, Ultra-Low Latency Media Streaming Server in Go**

*A high-performance, lightweight, drop-in replacement for Lavalink powering modern Discord bots.*

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Docker Ready](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![Lavalink Compatibility](https://img.shields.io/badge/Lavalink-v4_Compatible-5865F2?style=flat-square&logo=discord)](https://github.com/lavalink-devs/Lavalink)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

---

[Key Features](#-key-features) •
[Architecture Comparison](#-architecture-comparison) •
[Quick Start](#-quick-start) •
[Bot Configuration](#-connecting-your-bot) •
[Web Dashboard](#-real-time-web-dashboard) •
[Documentation](#-documentation)

</div>

---

## ⚡ Overview

**Aurix Media Engine (AME)** is a state-of-the-art media audio streaming server engineered from the ground up in **Go**. Designed specifically for high-scale Discord bot applications, AME replaces resource-heavy JVM media servers with a native, concurrency-first streaming architecture.

### Why Aurix Media Engine?

* 📉 **80%+ Memory Cost Reduction**: Operates at **~20–40 MB RAM** per node (compared to **300 MB–1 GB+** for Java-based Lavalink).
* ⏱️ **Zero Audio Stuttering**: Sub-millisecond concurrent garbage collection eliminates Stop-The-World (STW) audio frame drops.
* 🚀 **Instant Binary Startup**: Boots in **< 100 ms** with single static binary deployments or lightweight **25 MB Docker containers**.
* 🔄 **100% Lavalink v4 Protocol Compatible**: Works seamlessly with existing bot libraries (`lavalink.js`, `wavelink`, `poru`, `erela.js`, `disnake`).

---

## 🔥 Key Features

- **⚡ Lightweight Concurrency**: Driven by Go goroutines capable of handling tens of thousands of concurrent voice streams.
- **🔊 Discord Voice Gateway v8**: Full UDP STUN IP discovery, RTP packet framing, and modern encryption ciphers (`aead_xchacha20_poly1305_rtpsize`, `xsalsa20_poly1305`).
- **🎵 Multi-Source Audio Engine**:
  - **YouTube & YouTube Music**: Powered by `yt-dlp` streaming pipes.
  - **Direct Audio Links**: Direct `.mp3`, `.ogg`, `.flac`, `.wav`, `.m3u8` streams.
  - **Local Media Files**: Safe filesystem audio resolution with path traversal security.
  - **Spotify Link Metadata**: Resolves track and playlist links.
- **🎛️ Digital Signal Processing (DSP) Filters**: 15-band Equalizer (Bass Boost, Treble), Volume Gain, Timescale, and Nightcore filters.
- **📊 Embedded Real-time Web Dashboard**: Built-in glassmorphism dark-mode telemetry dashboard at `http://localhost:2333/dashboard`.
- **🐳 Multi-Stage Docker Build**: Production-ready container image bundled with FFmpeg runtime.

---

## 📊 Architecture Comparison

| Feature / Metric | **Aurix Media Engine (AME)** | **Lavalink (v4)** | **Andes** |
| :--- | :---: | :---: | :---: |
| **Language** | **Go** ⚡ | Java (JVM) ☕ | Go |
| **Runtime Footprint (Idle)** | **~20 - 40 MB** 🟢 | 300 MB - 1 GB 🔴 | ~30 - 50 MB |
| **Startup Time** | **< 100 ms** ⚡ | 5 - 15 seconds 🔴 | < 100 ms |
| **GC Latency Spikes** | **< 1 ms** (Tri-color) | 10 - 100 ms (Jitter) | < 1 ms |
| **Binary Size** | **~25 MB** | ~150 MB (with JRE) | ~30 MB |
| **Lavalink v4 Protocol API** | ✅ **Full Compatibility** | ✅ Native | ❌ Partial v3 |
| **Built-in Web Dashboard** | ✅ **Included (`/dashboard`)** | ❌ None | ❌ None |

---

## 🚀 Quick Start

### Option 1: Running Locally (No Docker Required)

#### Prerequisites
* [Go 1.22+](https://go.dev/dl/)
* [FFmpeg](https://ffmpeg.org/) (installed and added to system `PATH`)

```powershell
# Clone the repository
git clone https://github.com/sebastianjamesog/Aurix-Media-Engine.git
cd Aurix-Media-Engine

# Build the binary
go build -o server.exe ./cmd/server

# Start the server
.\server.exe
```

The gateway will start on **`http://localhost:2333`** with default password **`youshallnotpass`**.

---

### Option 2: Running via Docker Compose 🐳

```bash
docker-compose up -d
```

---

## 🤖 Connecting Your Bot

Plug these node credentials directly into any **Lavalink v4 compatible bot library**:

```yaml
Host: localhost (or your server IP)
Port: 2333
Password: youshallnotpass
Secure: false
```

### Node.js Example (`discord.js` + `lavalink.js`)

```javascript
const { Client, GatewayIntentBits } = require('discord.js');
const { Lavalink } = require('lavalink.js');

const client = new Client({
    intents: [
        GatewayIntentBits.Guilds,
        GatewayIntentBits.GuildVoiceStates,
        GatewayIntentBits.GuildMessages,
        GatewayIntentBits.MessageContent
    ]
});

client.on('clientReady', () => {
    console.log(`Bot logged in as ${client.user.tag}`);
    
    // Connect to Aurix Media Engine Node
    client.lavalink = new Lavalink({
        host: 'localhost',
        port: 2333,
        password: 'youshallnotpass',
        userId: client.user.id,
        numShards: 1
    });

    client.lavalink.on('connected', () => {
        console.log('[AME] Connected to Aurix Media Engine node successfully!');
    });
});

client.login('YOUR_DISCORD_BOT_TOKEN');
```

---

### Python Example (`discord.py` + `wavelink`)

```python
import discord
from discord.ext import commands
import wavelink

intents = discord.Intents.default()
intents.message_content = True
bot = commands.Bot(command_prefix="!", intents=intents)

@bot.event
async def on_ready():
    print(f"Logged in as {bot.user}")
    node = wavelink.Node(
        identifier="aurix-node",
        uri="http://localhost:2333",
        password="youshallnotpass"
    )
    await wavelink.Pool.connect(nodes=[node], client=bot)

bot.run("YOUR_DISCORD_BOT_TOKEN")
```

---

## 📊 Real-time Web Dashboard

Open **[http://localhost:2333/dashboard](http://localhost:2333/dashboard)** in your browser to monitor live node telemetry:

* Active Players & Playing Channels
* Engine Uptime & Monitored Memory Allocation
* Active Voice Connections & Audio Stream Health

---

## 📚 Documentation & Specifications

* 📘 [Quick Start Guide](docs/quickstart.md)
* 📙 [API & WebSocket Reference](docs/api-reference.md)
* 📗 [Discord Voice Protocol Spec](research/discord-voice.md)
* 📕 [Audio Pipeline Spec](research/audio-pipeline.md)
* 📔 [System Architecture](research/architecture.md)

---

## 🏷️ Tags & Keywords

`discord-bot` • `lavalink-alternative` • `audio-streaming-server` • `go` • `golang` • `opus` • `discord-voice` • `music-bot` • `lavalink-v4` • `yt-dlp` • `ffmpeg` • `high-performance` • `microservices`

---

## 📄 License

Distributed under the MIT License. See [`LICENSE`](LICENSE) for details.
