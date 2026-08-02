# Architecture & Performance Comparison

## Overview

This document provides a comparative analysis of **Aurix Media Engine (AME)** against established media server solutions in the Discord bot ecosystem (primarily **Lavalink**, **Andes**, and **AudioLink**).

---

## 1. Feature Matrix

| Feature / Metric | **Aurix Media Engine (AME)** | **Lavalink (v4)** | **Andes** | **AudioLink** |
| :--- | :---: | :---: | :---: | :---: |
| **Primary Language** | **Go** | Java (JVM) | Go | C++ |
| **Runtime Footprint (Idle)** | **15 - 40 MB** | 300 MB - 1 GB | 20 - 50 MB | 10 - 30 MB |
| **Startup Time** | **< 100 ms** | 5 - 15 seconds | < 100 ms | < 50 ms |
| **Concurrency Primitive** | **Goroutines (2KB)** | OS Threads / Netty | Goroutines | OS Thread Pools |
| **GC Pause Latency** | **< 1 ms (Tri-color)** | 10 - 100 ms (Jitter risk)| < 1 ms | None (No GC) |
| **Modern Ciphers (AEAD XChaCha20)** | ✅ Built-in | ✅ Supported | 🟡 Limited | 🟡 Limited |
| **Native Plugin Architecture** | ✅ Go Plugins & Interfaces | ✅ Java JAR plugins | ❌ Hardcoded | ❌ Hardcoded |
| **Native Clustering & Failover** | ✅ Integrated (Redis/gRPC) | 🟡 External (Lavalink-Filter) | ❌ Single-node | ❌ Single-node |
| **Lavalink v4 API Compatibility** | ✅ Full Protocol Compatible | ✅ Native | ❌ Partial v3 | ❌ Custom |

---

## 2. Key Advantages of AME Over Lavalink

### 1. Memory Efficiency & Hosting Cost Reduction
* **Lavalink**: Requires the JVM (Java Virtual Machine). Even with G1GC or ZGC optimizations, running 10 Lavalink nodes consumes **8 - 16 GB of RAM**, translating to high cloud server bills ($80–$150/month).
* **AME**: Native Go binaries consume **~30MB of RAM** per node. You can run 10 AME nodes on a single **1 GB RAM VPS ($5/month)**, reducing hosting infrastructure costs by **over 80%**.

### 2. Elimination of Audio Stuttering (GC Pauses)
* Discord audio requires a 20ms UDP frame ticker. In Java, garbage collection sweeps can cause occasional **20–100ms Stop-The-World (STW)** pauses, causing audio frame drops and audible pops/clicks.
* Go’s concurrent collector runs concurrently with goroutines, ensuring frame pacer tickers maintain **sub-millisecond accuracy**.

### 3. Horizontal Scaling & High Availability (HA)
* Lavalink nodes operate completely independently without built-in inter-node state sharing. If a Lavalink node goes down, all active music players on that node disconnect and die.
* AME is built with a **native distributed clustering layer** (`internal/cluster`). If Node A experiences a hardware failure, AME automatically re-assigns player state to Node B and reconnects Discord voice sockets transparently.

### 4. Single-Binary Container Deployment
* Docker images for Lavalink require a heavy Java JRE base image (~200MB - 400MB image size).
* AME builds into a scratch / Alpine single binary image of **~15MB - 30MB**, speeding up CI/CD pipelines and deployment times.
