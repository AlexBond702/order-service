package section

type Monitor struct {
	Sentry      MonitorSentry
	Prometheus  MonitorPrometheus
	LogLevel    string `default:"trace" split_word:"true"`
	Environment string `default:"dev"`
}
type MonitorPrometheus struct {
	Enabled bool `default:"false"`
}
type MonitorSentry struct {
	Enabled bool   `default:"false"`
	DSN     string `default:""`
}
