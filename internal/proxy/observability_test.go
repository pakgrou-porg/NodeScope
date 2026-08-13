package proxy

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestOperationalEventUsesOnlyApprovedMetadataFields(t *testing.T) {
	allowed := map[string]bool{
		"OccurredAt": true, "RouteID": true, "Model": true, "ClientID": true, "BackendID": true,
		"StatusCode": true, "Streaming": true, "TTFTMilliseconds": true, "DurationMilliseconds": true,
		"PromptTokens": true, "OutputTokens": true, "TotalTokens": true, "Outcome": true,
	}
	typeOfEvent := reflect.TypeFor[OperationalEvent]()
	if typeOfEvent.NumField() != len(allowed) {
		t.Fatalf("OperationalEvent field count = %d, want %d", typeOfEvent.NumField(), len(allowed))
	}
	for index := 0; index < typeOfEvent.NumField(); index++ {
		if !allowed[typeOfEvent.Field(index).Name] {
			t.Fatalf("OperationalEvent exposes non-allowlisted field %q", typeOfEvent.Field(index).Name)
		}
	}
}

func TestMetadataOnlyFanoutDeliversSameSafeEventToEveryOperationalSurface(t *testing.T) {
	requestLog := &memorySink{}
	tracer := &memorySink{}
	auditor := &memorySink{}
	supportExport := &memorySink{}
	event := OperationalEvent{OccurredAt: time.Unix(1, 0), RouteID: "route-a", Model: "model-a", ClientID: "client-a", BackendID: "route-a:primary", Outcome: "completed"}
	fanout := MetadataOnlyFanout{RequestLog: requestLog, Tracer: tracer, Auditor: auditor, SupportExport: supportExport}
	if err := fanout.ObserveProxy(context.Background(), event); err != nil {
		t.Fatalf("ObserveProxy() error = %v", err)
	}
	for name, sink := range map[string]*memorySink{"request log": requestLog, "tracer": tracer, "auditor": auditor, "support export": supportExport} {
		if len(sink.events) != 1 || sink.events[0] != event {
			t.Fatalf("%s received %#v, want %#v", name, sink.events, event)
		}
	}
}

type memorySink struct{ events []OperationalEvent }

func (sink *memorySink) WriteOperationalEvent(_ context.Context, event OperationalEvent) error {
	sink.events = append(sink.events, event)
	return nil
}
