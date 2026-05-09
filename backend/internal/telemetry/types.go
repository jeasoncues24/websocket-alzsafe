package telemetry

import (
	"wsapi/internal/domain"
)

const ServicePrefix = "/api/service/v1/"

type Config struct {
	BufferSize  int
	FlushSecs   int
	BatchSize   int
	Enabled     bool
}

func DefaultConfig() Config {
	return Config{
		BufferSize: 1000,
		FlushSecs:  5,
		BatchSize:  100,
		Enabled:    true,
	}
}

type Store interface {
	Record(event *domain.TelemetryEvent) error
	Flush() error
	Close() error
}

func ExtractContractName(path string) string {
	if len(path) < len(ServicePrefix) {
		return "unknown"
	}
	rest := path[len(ServicePrefix):]
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' || rest[i] == '?' || rest[i] == '#' {
			return rest[:i]
		}
	}
	return rest
}
