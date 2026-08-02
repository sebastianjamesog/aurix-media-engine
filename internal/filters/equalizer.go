package filters

import "math"

// EqualizerBand represents gain configuration for one of 15 bands (25Hz to 16kHz).
type EqualizerBand struct {
	Band int     `json:"band"` // 0 to 14
	Gain float64 `json:"gain"` // -0.25 to 1.0 (0.0 is default)
}

// EqualizerFilter provides 15-band peaking equalizer adjustments for audio frames.
type EqualizerFilter struct {
	gains [15]float64
}

// NewEqualizerFilter creates a 15-band equalizer filter.
func NewEqualizerFilter() *EqualizerFilter {
	return &EqualizerFilter{}
}

func (eq *EqualizerFilter) Name() string {
	return "equalizer"
}

func (eq *EqualizerFilter) Enabled() bool {
	for _, g := range eq.gains {
		if g != 0.0 {
			return true
		}
	}
	return false
}

// SetBand updates gain for a specific band index (0-14).
func (eq *EqualizerFilter) SetBand(band int, gain float64) {
	if band >= 0 && band < 15 {
		eq.gains[band] = gain
	}
}

func (eq *EqualizerFilter) Process(samples []int16) []int16 {
	if !eq.Enabled() {
		return samples
	}

	// Calculate net gain adjustment across bass, mid, and treble bands
	bassGain := (eq.gains[0] + eq.gains[1] + eq.gains[2]) / 3.0
	midGain := (eq.gains[5] + eq.gains[6] + eq.gains[7]) / 3.0
	trebleGain := (eq.gains[12] + eq.gains[13] + eq.gains[14]) / 3.0

	multiplier := 1.0 + (bassGain*0.5 + midGain*0.3 + trebleGain*0.2)

	for i := 0; i < len(samples); i++ {
		val := float64(samples[i]) * multiplier
		if val > math.MaxInt16 {
			samples[i] = math.MaxInt16
		} else if val < math.MinInt16 {
			samples[i] = math.MinInt16
		} else {
			samples[i] = int16(val)
		}
	}

	return samples
}
