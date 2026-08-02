# Aurix Media Engine — Quick Start Guide

Welcome to **Aurix Media Engine (AME)**, a high-performance distributed media streaming server written in Go.

---

## 1. Prerequisites

* **Go 1.22+** installed on your system.
* **FFmpeg** installed and accessible via system `PATH` (for audio decoding).

---

## 2. Running Locally

### Step 1: Build the Executable
In the workspace root directory, run:

```powershell
go build -o server.exe ./cmd/server
```

### Step 2: Start the Server
```powershell
.\server.exe
```

The gateway will start on **`http://localhost:2333`** with the default password `youshallnotpass`.

---

## 3. Web Dashboard

Open **[http://localhost:2333/dashboard](http://localhost:2333/dashboard)** in your browser to monitor real-time CPU, RAM, active players, and engine statistics.

---

## 4. Connecting Your Discord Bot

Configure your bot's Lavalink client node settings:

```yaml
Host: localhost
Port: 2333
Password: youshallnotpass
Secure: false
```

### Supported Client Libraries
* **Node.js**: `lavalink.js`, `poru`, `erela.js`
* **Python**: `wavelink`, `mafic`, `pomice`
* **Go**: Native `aurix/sdk/go` client
