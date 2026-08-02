# Audio Pipeline Specification

## Overview

The **Audio Pipeline** in Aurix Media Engine (AME) is responsible for taking raw audio streams from providers (YouTube, HTTP streams, Spotify, Local files), decoding them into raw PCM samples, running optional real-time Digital Signal Processing (DSP) audio filters, encoding the resulting audio into 48kHz Stereo Opus frames, and delivering them to the packet pacer.

---

## 1. High-Level Dataflow

```
               ┌───────────────────────┐
               │    Track Streamer     │ (HTTP / Local File / Process Pipe)
               └───────────┬───────────┘
                           │ Raw Encoded Audio Stream (WebM / MP3 / AAC / FLAC)
                           ▼
               ┌───────────────────────┐
               │     Audio Decoder     │ (FFmpeg / Native Go Decoders)
               └───────────┬───────────┘
                           │ Interleaved 48kHz Stereo 16-bit PCM SAMPLES
                           ▼
               ┌───────────────────────┐
               │  DSP Filter Chain     │ (Volume, Equalizer, Timescale, Nightcore)
               └───────────┬───────────┘
                           │ Filtered 48kHz Stereo 16-bit PCM SAMPLES
                           ▼
               ┌───────────────────────┐
               │     Opus Encoder      │ (48kHz, 2 Channels, 20ms Frame = 960 samples/ch)
               └───────────┬───────────┘
                           │ 20ms Encoded Opus Packets
                           ▼
               ┌───────────────────────┐
               │  Frame Pacer RingBuf  │ (20ms High-Precision Ticker)
               └───────────┬───────────┘
                           │ Encrypted UDP Packet
                           ▼
               ┌───────────────────────┐
               │    Discord UDP Voice  │
               └───────────────────────┘
```

---

## 2. Core Audio Parameters

To meet Discord's strict voice input specifications:

* **Sample Rate**: 48,000 Hz (48 kHz)
* **Channels**: 2 (Stereo)
* **Bit Depth**: 16-bit signed integer (Little Endian PCM `pcm_s16le`) or 32-bit floating point (`pcm_f32le` for DSP operations)
* **Frame Duration**: 20 milliseconds (ms)
* **Samples per 20ms Frame**: 
  $$\text{Samples per channel} = 48,000 \times 0.020 = 960 \text{ samples}$$
  $$\text{Total samples (Stereo)} = 960 \times 2 = 1,920 \text{ samples}$$
  $$\text{Raw PCM Frame Size (16-bit)} = 1,920 \times 2 \text{ bytes} = 3,840 \text{ bytes}$$

---

## 3. Audio Ingestion & Decoding Strategy

### Decoder Architecture
AME abstracts decoding behind an `AudioDecoder` Go interface:

```go
type AudioDecoder interface {
    // ReadPCM reads up to len(pcmBuf) S16LE stereo samples.
    ReadPCM(pcmBuf []int16) (n int, err error)
    // Seek seeks to a given timestamp offset.
    Seek(offset time.Duration) error
    // Close cleans up decoder processes and memory.
    Close() error
}
```

### FFmpeg Integration Strategy
* For universal codec support (MP3, AAC, Ogg, FLAC, WebM, M3U8 HLS streams), AME uses standard OS FFmpeg pipes (`exec.Command`) or `libav*` CGO bindings.
* Command invocation format:
  ```bash
  ffmpeg -i <stream_url> -f s16le -ar 48000 -ac 2 pipe:1
  ```

---

## 4. Digital Signal Processing (DSP) & Filter Graph

Audio filters operate directly on PCM sample buffers prior to Opus encoding.

### Standard Supported Filters
1. **Volume**: Linear scalar multiplication:
   $$y[i] = x[i] \times \text{volume\_factor}$$
2. **Equalizer (15-band)**: Frequency band gains from 25 Hz to 16 kHz.
3. **Timescale / Speed / Pitch**: Dynamic rate and pitch conversion (using SoundTouch algorithm / phase vocoder).
4. **Nightcore**: Pitch increase (+30%) + Speed increase (+20%).
5. **Karaoke**: Center-channel voice cancellation:
   $$\text{mono} = \frac{L - R}{2}$$
6. **Rotation / 3D Audio**: Panning audio across stereo channels using low-frequency sine oscillation.

---

## 5. Opus Encoding & Frame Pacing

### Opus Encoder Settings
* **Application**: `voip` (optimizes latency) or `audio` (high-fidelity music). Default: `audio`.
* **Bitrate**: Configurable from 8 kbps to 320 kbps (Default: 64 kbps or 128 kbps).
* **VBR**: Variable Bitrate enabled.
* **Complexity**: 10 (maximum quality).

### Precision Ticker (Frame Pacer)
Because standard Go `time.Sleep` can fluctuate under OS thread scheduling, AME uses a dedicated **High-Precision Ring Buffer Pacer**:

```
        Ring Buffer (Pre-buffered Opus Frames: ~200ms / 10 frames)
  ┌────┬────┬────┬────┬────┬────┬────┬────┬────┬────┐
  │ F1 │ F2 │ F3 │ F4 │ F5 │ F6 │ F7 │ F8 │ F9 │F10 │
  └────┴────┴────┴────┴────┴────┴────┴────┴────┴────┘
    ▲                                       ▲
    │ Read by UDP Pacer every 20ms          │ Written by Pipeline Goroutine
```

* **Drift Compensation**: Tracks actual monotonic time vs. expected frame count to adjust ticker intervals dynamically if a packet send is delayed by OS context switches.
