package filters

import "testing"

func TestVolumeFilter(t *testing.T) {
	filter := NewVolumeFilter(0.5)

	samples := []int16{1000, -2000, 30000}
	out := filter.Process(samples)

	if out[0] != 500 {
		t.Errorf("expected 500, got %d", out[0])
	}
	if out[1] != -1000 {
		t.Errorf("expected -1000, got %d", out[1])
	}
	if out[2] != 15000 {
		t.Errorf("expected 15000, got %d", out[2])
	}
}

func TestEqualizerFilter(t *testing.T) {
	eq := NewEqualizerFilter()
	if eq.Enabled() {
		t.Errorf("expected default EQ filter to be disabled")
	}

	eq.SetBand(0, 0.5) // Boost bass
	if !eq.Enabled() {
		t.Errorf("expected EQ filter to be enabled after setting gain")
	}

	samples := []int16{1000, -1000}
	out := eq.Process(samples)

	if out[0] <= 1000 {
		t.Errorf("expected boosted sample > 1000, got %d", out[0])
	}
}
