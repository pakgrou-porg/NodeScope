package proxy

import (
	"reflect"
	"testing"
)

func TestMetadataOnlyEventFieldAllowlists(t *testing.T) {
	assertExactFieldNames(t, reflect.TypeOf(UsageEvent{}), []string{
		"OccurredAt",
		"RouteID",
		"Model",
		"ClientID",
		"BackendID",
		"StatusCode",
		"Streaming",
		"TTFTMilliseconds",
		"DurationMilliseconds",
		"PromptTokens",
		"OutputTokens",
		"TotalTokens",
		"Outcome",
	})
	assertExactFieldNames(t, reflect.TypeOf(OperationalEvent{}), []string{
		"OccurredAt",
		"RouteID",
		"Model",
		"ClientID",
		"BackendID",
		"StatusCode",
		"Streaming",
		"TTFTMilliseconds",
		"DurationMilliseconds",
		"PromptTokens",
		"OutputTokens",
		"TotalTokens",
		"Outcome",
	})
}

func assertExactFieldNames(t *testing.T, eventType reflect.Type, want []string) {
	t.Helper()
	if eventType.NumField() != len(want) {
		t.Fatalf("%s has %d fields; want exactly %d metadata-only fields", eventType.Name(), eventType.NumField(), len(want))
	}
	for index, name := range want {
		if got := eventType.Field(index).Name; got != name {
			t.Fatalf("%s field %d = %q; want %q", eventType.Name(), index, got, name)
		}
	}
}
