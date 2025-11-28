package api

import "github.com/samber/lo"

// ProfilingEvent represents the type of event to profile.
type ProfilingEvent string

const (
	Cpu         ProfilingEvent = "cpu"          // Cpu represents CPU profiling events.
	Alloc       ProfilingEvent = "alloc"        // Alloc represents memory allocation events.
	Lock        ProfilingEvent = "lock"         // Lock represents lock contention events.
	CacheMisses ProfilingEvent = "cache-misses" // CacheMisses represents CPU cache miss events.
	Wall        ProfilingEvent = "wall"         // Wall represents wall-clock time events.
	Itimer      ProfilingEvent = "itimer"       // Itimer represents interval timer events.
	Ctimer      ProfilingEvent = "ctimer"       // Ctimer represents CPU timer events.
)

var (
	// supportedEvents contains all supported profiling event types.
	supportedEvents = []ProfilingEvent{Cpu, Alloc, Lock, CacheMisses, Wall, Itimer, Ctimer}
)

// AvailableEvents returns the list of all supported profiling events.
func AvailableEvents() []ProfilingEvent {
	return supportedEvents
}

// IsSupportedEvent checks if the given event string is a supported profiling event.
// It returns true if the event is in the list of available events.
func IsSupportedEvent(event string) bool {
	return lo.Contains(AvailableEvents(), ProfilingEvent(event))
}
