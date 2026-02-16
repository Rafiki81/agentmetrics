package cli

import (
	"github.com/Rafiki81/libagentmetrics/agent"
	"github.com/Rafiki81/libagentmetrics/config"
	"github.com/Rafiki81/libagentmetrics/monitor"
)

type scanRuntime struct {
	cfg      *config.Config
	registry *agent.Registry
	detector *agent.Detector
}

type tokenCollector interface {
	Collect([]agent.Instance)
}

type gitCollector interface {
	Collect(*agent.Instance)
}

type sessionCollector interface {
	Collect(*agent.Instance)
}

var newTokenCollector = func() tokenCollector {
	return monitor.NewTokenMonitor()
}

var newGitCollector = func() gitCollector {
	return monitor.NewGitMonitor()
}

var newSessionCollector = func() sessionCollector {
	return monitor.NewSessionMonitor()
}

func newScanRuntime() *scanRuntime {
	cfg := config.Load()
	registry := agent.NewRegistry()
	detector := agent.NewDetector(registry, cfg)

	return &scanRuntime{
		cfg:      cfg,
		registry: registry,
		detector: detector,
	}
}

func (r *scanRuntime) scan() ([]agent.Instance, error) {
	return r.detector.Scan()
}

func collectTokenMetrics(agents []agent.Instance) {
	tokenMon := newTokenCollector()
	tokenMon.Collect(agents)
}

func collectGitAndSessionMetrics(agents []agent.Instance) {
	gitMon := newGitCollector()
	sessionMon := newSessionCollector()
	for i := range agents {
		gitMon.Collect(&agents[i])
		sessionMon.Collect(&agents[i])
	}
}

func collectSessionMetrics(agents []agent.Instance) {
	sessionMon := newSessionCollector()
	for i := range agents {
		sessionMon.Collect(&agents[i])
	}
}
