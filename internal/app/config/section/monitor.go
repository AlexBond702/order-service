package section

type Monitor struct {
	Prometheus  MonitorPrometheus
	LogLevel    string `default:"trace" split_word:"true"`
	Environment string `default:"dev"`
}
type MonitorPrometheus struct {
	Enabled bool `default:"false"`
}
