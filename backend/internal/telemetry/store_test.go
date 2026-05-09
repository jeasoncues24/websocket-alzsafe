package telemetry

import (
	"sort"
	"testing"
	"time"

	"wsapi/internal/domain"
)

func TestPercentileInt(t *testing.T) {
	cases := []struct {
		name   string
		vals   []int
		p      float64
		expect float64
	}{
		{"vacia", []int{}, 0.50, 0},
		{"un elemento p50", []int{100}, 0.50, 100},
		{"cinco elementos p50", []int{10, 20, 30, 40, 50}, 0.50, 30},
		{"cinco elementos p95", []int{10, 20, 30, 40, 50}, 0.95, 50},
		{"cinco elementos p99", []int{10, 20, 30, 40, 50}, 0.99, 50},
		{"diez elementos p50", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 0.50, 5},
		{"diez elementos p90", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 0.90, 9},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sorted := make([]int, len(tc.vals))
			copy(sorted, tc.vals)
			sort.Ints(sorted)
			got := percentileInt(sorted, tc.p)
			if got != tc.expect {
				t.Fatalf("percentileInt(%v, %.2f) = %.0f, want %.0f", tc.vals, tc.p, got, tc.expect)
			}
		})
	}
}

func TestAggregateToBuckets_EmptyBatch(t *testing.T) {
	buckets := aggregateToBuckets([]*domain.TelemetryEvent{})
	if len(buckets) != 0 {
		t.Fatalf("expected 0 buckets, got %d", len(buckets))
	}
}

func TestAggregateToBuckets_SkipsZeroApiKey(t *testing.T) {
	events := []*domain.TelemetryEvent{
		{ApiKeyID: 0, ContractName: "mensajes", StatusCode: 200, LatencyMS: 50, CreatedAt: time.Now()},
	}
	buckets := aggregateToBuckets(events)
	if len(buckets) != 0 {
		t.Fatalf("expected 0 buckets for zero api_key_id, got %d", len(buckets))
	}
}

func TestAggregateToBuckets_SingleBucket(t *testing.T) {
	base := time.Date(2026, 5, 9, 12, 30, 15, 0, time.UTC)
	events := []*domain.TelemetryEvent{
		{ApiKeyID: 1, ContractName: "mensajes", StatusCode: 200, LatencyMS: 100, CreatedAt: base},
		{ApiKeyID: 1, ContractName: "mensajes", StatusCode: 200, LatencyMS: 200, CreatedAt: base.Add(20 * time.Second)},
		{ApiKeyID: 1, ContractName: "mensajes", StatusCode: 500, LatencyMS: 300, CreatedAt: base.Add(40 * time.Second)},
	}
	buckets := aggregateToBuckets(events)
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	b := buckets[0]
	if b.RequestCount != 3 {
		t.Errorf("RequestCount = %d, want 3", b.RequestCount)
	}
	if b.SuccessCount != 2 {
		t.Errorf("SuccessCount = %d, want 2", b.SuccessCount)
	}
	if b.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", b.ErrorCount)
	}
	if b.BucketMin != truncateToMinute(base) {
		t.Errorf("BucketMin = %v, want %v", b.BucketMin, truncateToMinute(base))
	}
}

func TestAggregateToBuckets_MultipleBuckets(t *testing.T) {
	min1 := time.Date(2026, 5, 9, 12, 30, 0, 0, time.UTC)
	min2 := time.Date(2026, 5, 9, 12, 31, 0, 0, time.UTC)
	events := []*domain.TelemetryEvent{
		{ApiKeyID: 1, ContractName: "mensajes", StatusCode: 200, LatencyMS: 50, CreatedAt: min1},
		{ApiKeyID: 1, ContractName: "mensajes", StatusCode: 200, LatencyMS: 60, CreatedAt: min2},
		{ApiKeyID: 1, ContractName: "sesion", StatusCode: 200, LatencyMS: 10, CreatedAt: min1},
	}
	buckets := aggregateToBuckets(events)
	if len(buckets) != 3 {
		t.Fatalf("expected 3 buckets (2 min×contrato), got %d", len(buckets))
	}
}

func TestAggregateToBuckets_TruncatesToMinute(t *testing.T) {
	base := time.Date(2026, 5, 9, 12, 30, 0, 0, time.UTC)
	events := []*domain.TelemetryEvent{
		{ApiKeyID: 1, ContractName: "x", StatusCode: 200, LatencyMS: 1, CreatedAt: base.Add(5 * time.Second)},
		{ApiKeyID: 1, ContractName: "x", StatusCode: 200, LatencyMS: 2, CreatedAt: base.Add(59 * time.Second)},
	}
	buckets := aggregateToBuckets(events)
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket (mismo minuto), got %d", len(buckets))
	}
	if buckets[0].BucketMin != base {
		t.Errorf("BucketMin = %v, want %v", buckets[0].BucketMin, base)
	}
}
