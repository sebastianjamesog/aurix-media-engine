package filters

import "math"

// VolumeFilter adjusts the gain/volume of 16-bit PCM samples.
type VolumeFilter struct {
	factor float64 // 1.0 = 100% volume
}

// NewVolumeFilter creates a volume gain DSP filter.
func NewVolumeFilter(volume float64) *VolumeFilter {
	return &VolumeFilter{factor: volume}
}

func (v *VolumeFilter) Name() string {
	return "volume"
}

func (v *VolumeFilter) Enabled() bool {
	return v.factor != 1.0
}

func (v *VolumeFilter) SetVolume(volume float64) {
	v.factor = volume
}

func (v *VolumeFilter) Process(samples []int16) []int16 {
	if v.factor == 1.0 {
		return samples
	}

	for i := 0; i < len(samples); i++ {
		val := float64(samples[i]) * v.factor
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
