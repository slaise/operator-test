package health

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	k8sclock "k8s.io/apimachinery/pkg/util/clock"
)

type HealthCheck interface {
	// Check implements healthz.Checker
	Check(*http.Request) error
	// Trigger update last run time
	Trigger()
}

type healthCheck struct {
	clock     k8sclock.PassiveClock
	tolerance time.Duration
	log       logr.Logger
	lastTime  atomic.Value
}

var _ HealthCheck = (*healthCheck)(nil)

// NewHealthCheck creates a new HealthCheck that will calculate the time inactive
// based on the provided clock and configuration.
func NewHealthCheck(clock k8sclock.PassiveClock,
	tolerance time.Duration,
	log logr.Logger) HealthCheck {

	answer := &healthCheck{
		clock:     clock,
		tolerance: tolerance,
		log:       log.WithName("HealthCheck"),
	}
	answer.Trigger()
	return answer
}

func (a *healthCheck) Trigger() {
	a.lastTime.Store(a.clock.Now())
}

func (a *healthCheck) Check(_ *http.Request) error {
	lastActivity := a.lastTime.Load().(time.Time)
	if a.clock.Now().After(lastActivity.Add(a.tolerance)) {
		err := fmt.Errorf("last activity more than %s ago (%s)",
			a.tolerance, lastActivity.Format(time.RFC3339))
		a.log.Error(err, "Failing activity health check")
		return err
	}
	return nil
}
