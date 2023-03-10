package api

import "github.com/samber/lo"

type ProfilingEvent string

const (
	Cpu         ProfilingEvent = "cpu"
	Alloc       ProfilingEvent = "alloc"
	Lock        ProfilingEvent = "lock"
	CacheMisses ProfilingEvent = "cache-misses"
	Wall        ProfilingEvent = "wall"
	Itimer      ProfilingEvent = "itimer"
)

var (
	supportedEvents = []ProfilingEvent{Cpu, Alloc, Lock, CacheMisses, Wall, Itimer}
)

func AvailableEvents() []ProfilingEvent {
	return supportedEvents
}

func IsSupportedEvent(event string) bool {
	return lo.Contains(AvailableEvents(), ProfilingEvent(event))
}
