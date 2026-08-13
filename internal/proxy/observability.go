package proxy

import (
	"context"
	"errors"
)

// OperationalSink is the narrow adapter contract for request logs, traces,
// audit records, and support exports. It receives an OperationalEvent rather
// than a request, response, header collection, or inference payload.
type OperationalSink interface {
	WriteOperationalEvent(context.Context, OperationalEvent) error
}

// MetadataOnlyFanout delivers one safe operational event to optional logging,
// tracing, audit, and support-export sinks. It gives integrations no route to
// observe prompt text, response text, tool arguments, credentials, or bytes.
type MetadataOnlyFanout struct {
	RequestLog    OperationalSink
	Tracer        OperationalSink
	Auditor       OperationalSink
	SupportExport OperationalSink
}

func (fanout MetadataOnlyFanout) ObserveProxy(ctx context.Context, event OperationalEvent) error {
	var failures []error
	for _, sink := range []OperationalSink{fanout.RequestLog, fanout.Tracer, fanout.Auditor, fanout.SupportExport} {
		if sink == nil {
			continue
		}
		if err := sink.WriteOperationalEvent(ctx, event); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
