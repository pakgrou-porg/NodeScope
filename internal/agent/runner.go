package agent

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/domain"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

type TelemetrySender interface {
	Send(context.Context, telemetry.Envelope) error
}

type Runner struct {
	config      Config
	collectors  []Collector
	sender      TelemetrySender
	state       *SequenceStore
	now         func() time.Time
	startedAt   time.Time
	retryPolicy RetryPolicy
	random      func() float64
	sleep       func(context.Context, time.Duration) error
	warn        func(string)
	reportError func(error)
}

func NewRunner(config Config, collectors []Collector, sender TelemetrySender, state *SequenceStore) (*Runner, error) {
	if sender == nil || state == nil {
		return nil, fmt.Errorf("agent sender and sequence state are required")
	}
	if len(collectors) == 0 {
		return nil, fmt.Errorf("at least one collector is required")
	}
	policy := DefaultRetryPolicy()
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	startedAt := time.Now()
	return &Runner{
		config:      config,
		collectors:  collectors,
		sender:      sender,
		state:       state,
		now:         time.Now,
		startedAt:   startedAt,
		retryPolicy: policy,
		random:      nil,
		sleep:       sleepWithContext,
		warn: func(message string) {
			log.Printf("nodescope-agent warning: %s", message)
		},
		reportError: func(err error) {
			// Error values can contain transport detail. Retain only their type in
			// process logs so periodic observability does not disclose endpoints.
			log.Printf("nodescope-agent collection cycle failed: %T", err)
		},
	}, nil
}

func (runner *Runner) CollectOnce(ctx context.Context) error {
	observedAt := runner.now().UTC()
	samples := make([]telemetry.Sample, 0, 128)
	containers := make([]telemetry.ContainerInventory, 0, 64)
	for _, collector := range runner.collectors {
		if inventoryCollector, ok := collector.(ContainerInventoryCollector); ok {
			collected, inventory, err := inventoryCollector.CollectContainerInventory(ctx, observedAt)
			if err != nil {
				runner.warn(fmt.Sprintf("collector %q did not provide container inventory", collector.Name()))
				samples = append(samples, collectorUnavailable(collector.Name(), observedAt))
				continue
			}
			samples = append(samples, collected...)
			containers = append(containers, inventory...)
			continue
		}
		collected, err := collector.Collect(ctx, observedAt)
		if err != nil {
			runner.warn(fmt.Sprintf("collector %q did not provide a reading", collector.Name()))
			samples = append(samples, collectorUnavailable(collector.Name(), observedAt))
			continue
		}
		samples = append(samples, collected...)
	}
	if len(samples) == 0 {
		return fmt.Errorf("all collectors returned no samples")
	}
	bootID, sequence, err := runner.state.Next()
	if err != nil {
		return err
	}
	checksum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", runner.config.AgentID, bootID, sequence)))
	envelope := telemetry.Envelope{
		SchemaVersion:        telemetry.CurrentSchemaVersion,
		Codec:                telemetry.CodecProtoZstd,
		AgentID:              runner.config.AgentID,
		HostID:               runner.config.HostID,
		BootID:               bootID,
		Sequence:             sequence,
		ObservedAt:           observedAt,
		MonotonicElapsedNano: uint64(time.Since(runner.startedAt).Nanoseconds()),
		SampleCount:          uint32(len(samples)),
		MetricValueCount:     uint32(len(samples)),
		UncompressedBytes:    1,
		CompressedBytes:      1,
		ChecksumSHA256:       fmt.Sprintf("%x", checksum),
		Samples:              samples,
		Containers:           containers,
	}
	return runner.sendWithRetry(ctx, envelope)
}

func (runner *Runner) collectAndReport(ctx context.Context) {
	if err := runner.CollectOnce(ctx); err != nil {
		runner.reportError(err)
	}
}

func (runner *Runner) sendWithRetry(ctx context.Context, envelope telemetry.Envelope) error {
	var lastErr error
	for attempt := 0; attempt < runner.retryPolicy.MaxAttempts; attempt++ {
		if err := runner.sender.Send(ctx, envelope); err == nil {
			return nil
		} else {
			lastErr = err
			if !isRetryable(err) {
				return err
			}
		}
		if attempt == runner.retryPolicy.MaxAttempts-1 {
			break
		}
		if err := runner.sleep(ctx, runner.retryPolicy.Delay(attempt, runner.random)); err != nil {
			return err
		}
	}
	return lastErr
}

func (runner *Runner) Run(ctx context.Context) error {
	runner.collectAndReport(ctx)
	ticker := time.NewTicker(runner.config.CollectionInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			runner.collectAndReport(ctx)
		}
	}
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func collectorUnavailable(name string, observedAt time.Time) telemetry.Sample {
	return telemetry.Sample{DeviceID: "collector:" + name, Metric: domain.MetricValue{Name: "collector.availability", Unit: "state", Quality: domain.QualityUnavailable, Source: "nodescope-agent", Semantics: "collector did not provide a reading", ObservedAt: observedAt}}
}
