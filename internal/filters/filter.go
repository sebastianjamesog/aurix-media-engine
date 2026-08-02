package filters

// AudioFilter defines the interface implemented by all audio DSP filters.
type AudioFilter interface {
	// Name returns the identifier of the filter.
	Name() string
	// Process operates in-place or returns processed 48kHz Stereo S16LE PCM samples.
	Process(samples []int16) []int16
	// Enabled returns whether the filter is active.
	Enabled() bool
}

// FilterChain manages a chain of sequential AudioFilters.
type FilterChain struct {
	filters []AudioFilter
}

// NewFilterChain creates a new empty FilterChain.
func NewFilterChain() *FilterChain {
	return &FilterChain{
		filters: make([]AudioFilter, 0),
	}
}

// Add appends a filter to the chain.
func (fc *FilterChain) Add(f AudioFilter) {
	fc.filters = append(fc.filters, f)
}

// Process passes samples through all active filters in the chain sequentially.
func (fc *FilterChain) Process(samples []int16) []int16 {
	out := samples
	for _, f := range fc.filters {
		if f.Enabled() {
			out = f.Process(out)
		}
	}
	return out
}
