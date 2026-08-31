package http

import "testing"

func TestMapV1EventType(t *testing.T) {
	cases := []struct {
		name  string
		event string
		data  map[string]any
		want  string
	}{
		{"qr", "qr-51999999999", nil, "qr"},
		{"connected", "active-51999999999", map[string]any{"isActive": true}, "connected"},
		{"disconnected", "active-51999999999", map[string]any{"isActive": false, "reason": "qr_timeout"}, "disconnected"},
		{
			name:  "client_outdated",
			event: "active-51999999999",
			data:  map[string]any{"isActive": false, "reason": "client_outdated"},
			want:  "client_outdated",
		},
		{"disconnected sin data", "active-51999999999", nil, "disconnected"},
		{"passthrough", "ping", nil, "ping"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := mapV1EventType(tc.event, tc.data); got != tc.want {
				t.Fatalf("mapV1EventType(%q, %v) = %q, want %q", tc.event, tc.data, got, tc.want)
			}
		})
	}
}
