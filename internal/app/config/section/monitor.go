package section

import "time"

type Monitor struct {
	Sentry        MonitorSentry
	Prometheus    MonitorPrometheus
	LogLevel      string               `default:"trace" split_word:"true"`
	Environment   string               `default:"dev"`
	OpenTelemetry MonitorOpenTelemetry `split_words:"true"`
}
type MonitorPrometheus struct {
	Enabled bool `default:"false"`
}
type MonitorSentry struct {
	Enabled bool   `default:"false"`
	DSN     string `default:""`
}
type MonitorOpenTelemetry struct {
	Enabled          bool          `default:"false"`
	Address          string        `default:""`
	MaxQueueSize     uint64        `default:"2048" split_words:"true"`
	MaxBatchSize     uint64        `default:"512" split_words:"true"`
	SendBatchTimeout time.Duration `default:"5s" split_words:"true"`
	ExportTimeout    time.Duration `default:"30s" split_words:"true"`
	SampleRatio      float64       `default:"1" split_words:"true"`
}
