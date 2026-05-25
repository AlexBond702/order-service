package section

type Monitor struct {
	LogLevel    string `default:"trace" split_word:"true"`
	Environment string `default:"dev"`
}
