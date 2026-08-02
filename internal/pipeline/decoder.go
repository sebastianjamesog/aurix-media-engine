package pipeline

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// FFmpegDecoder uses an ffmpeg pipe to decode audio streams to 48kHz Stereo S16LE PCM.
type FFmpegDecoder struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

// NewFFmpegDecoder launches an FFmpeg child process reading from an audio stream or URL.
func NewFFmpegDecoder(ctx context.Context, inputURL string) (*FFmpegDecoder, error) {
	// FFmpeg parameters: 48kHz sample rate, stereo (2 channels), 16-bit little-endian PCM
	args := []string{
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", inputURL,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1",
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe for ffmpeg: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg process: %w", err)
	}

	return &FFmpegDecoder{
		cmd:    cmd,
		stdout: stdout,
	}, nil
}

// ReadPCM reads raw S16LE stereo PCM bytes into buf.
// 20ms of audio = 3,840 bytes.
func (f *FFmpegDecoder) ReadPCM(buf []byte) (int, error) {
	return io.ReadFull(f.stdout, buf)
}

// Close terminates the FFmpeg process and cleans up pipes.
func (f *FFmpegDecoder) Close() error {
	if f.cmd != nil && f.cmd.Process != nil {
		_ = f.cmd.Process.Kill()
		return f.cmd.Wait()
	}
	return nil
}
