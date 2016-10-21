package trafficopswrapper

import (
	"fmt"
	"sync"

	to "github.com/apache/incubator-trafficcontrol/traffic_ops/client"
)

// ITrafficOpsSession provides an interface to the Traffic Ops client, so it may be wrapped or mocked.
type ITrafficOpsSession interface {
	CRConfigRaw(cdn string) ([]byte, error)
	TrafficMonitorConfigMap(cdn string) (*to.TrafficMonitorConfigMap, error)
	Set(session *to.Session)
}

// TrafficOpsSessionThreadsafe provides access to the Traffic Ops client safe for multiple goroutines. This fulfills the ITrafficOpsSession interface.
type TrafficOpsSessionThreadsafe struct {
	session **to.Session // pointer-to-pointer, because we're given a pointer from the Traffic Ops package, and we don't want to copy it.
	m       *sync.Mutex
}

// NewTrafficOpsSessionThreadsafe returns a new threadsafe TrafficOpsSessionThreadsafe wrapping the given `Session`.
func NewTrafficOpsSessionThreadsafe(s *to.Session) TrafficOpsSessionThreadsafe {
	return TrafficOpsSessionThreadsafe{&s, &sync.Mutex{}}
}

// CRConfigRaw returns the CRConfig from the Traffic Ops. This is safe for multiple goroutines.
func (s TrafficOpsSessionThreadsafe) CRConfigRaw(cdn string) ([]byte, error) {
	s.m.Lock()
	if s.session == nil || *s.session == nil {
		return nil, fmt.Errorf("nil session")
	}
	b, _, e := (*s.session).GetCRConfig(cdn)
	s.m.Unlock()
	return b, e
}

// TrafficMonitorConfigMap returns the Traffic Monitor config map from the Traffic Ops. This is safe for multiple goroutines.
func (s TrafficOpsSessionThreadsafe) TrafficMonitorConfigMap(cdn string) (*to.TrafficMonitorConfigMap, error) {
	s.m.Lock()
	if s.session == nil || *s.session == nil {
		return nil, fmt.Errorf("nil session")
	}
	d, e := (*s.session).TrafficMonitorConfigMap(cdn)
	s.m.Unlock()
	return d, e
}

// Set sets the internal Traffic Ops session. This is safe for multiple goroutines, being aware they will race.
func (s TrafficOpsSessionThreadsafe) Set(session *to.Session) {
	s.m.Lock()
	*s.session = session
	s.m.Unlock()
}
