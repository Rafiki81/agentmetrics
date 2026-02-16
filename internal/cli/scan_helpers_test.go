package cli

import (
	"testing"

	"github.com/Rafiki81/libagentmetrics/agent"
)

type fakeTokenCollector struct {
	called bool
	count  int
}

func (f *fakeTokenCollector) Collect(instances []agent.Instance) {
	f.called = true
	f.count = len(instances)
}

type fakeGitCollector struct {
	called int
}

func (f *fakeGitCollector) Collect(_ *agent.Instance) {
	f.called++
}

type fakeSessionCollector struct {
	called int
}

func (f *fakeSessionCollector) Collect(_ *agent.Instance) {
	f.called++
}

func TestCollectTokenMetricsCallsCollector(t *testing.T) {
	original := newTokenCollector
	defer func() { newTokenCollector = original }()

	fake := &fakeTokenCollector{}
	newTokenCollector = func() tokenCollector { return fake }

	agents := []agent.Instance{{}, {}}
	collectTokenMetrics(agents)

	if !fake.called {
		t.Fatalf("expected token collector to be called")
	}
	if fake.count != len(agents) {
		t.Fatalf("expected %d instances, got %d", len(agents), fake.count)
	}
}

func TestCollectGitAndSessionMetricsCallsEachPerAgent(t *testing.T) {
	origGit := newGitCollector
	origSession := newSessionCollector
	defer func() {
		newGitCollector = origGit
		newSessionCollector = origSession
	}()

	gitFake := &fakeGitCollector{}
	sessionFake := &fakeSessionCollector{}

	newGitCollector = func() gitCollector { return gitFake }
	newSessionCollector = func() sessionCollector { return sessionFake }

	agents := []agent.Instance{{}, {}, {}}
	collectGitAndSessionMetrics(agents)

	if gitFake.called != len(agents) {
		t.Fatalf("expected git collector called %d times, got %d", len(agents), gitFake.called)
	}
	if sessionFake.called != len(agents) {
		t.Fatalf("expected session collector called %d times, got %d", len(agents), sessionFake.called)
	}
}

func TestCollectSessionMetricsCallsEachPerAgent(t *testing.T) {
	original := newSessionCollector
	defer func() { newSessionCollector = original }()

	sessionFake := &fakeSessionCollector{}
	newSessionCollector = func() sessionCollector { return sessionFake }

	agents := []agent.Instance{{}, {}}
	collectSessionMetrics(agents)

	if sessionFake.called != len(agents) {
		t.Fatalf("expected session collector called %d times, got %d", len(agents), sessionFake.called)
	}
}
