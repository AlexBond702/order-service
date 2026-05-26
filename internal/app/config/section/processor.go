package section

import "time"

type (
	Processor struct {
		WebServer ProcessorWebServer `required:"true" split_words:"true"`
	}
	ProcessorWebServer struct {
		ListenPort uint32        `required:"true" default:"9020" split_words:"true"`
		Host       string        `required:"true" default:"localhost"`
		Timeout    time.Duration `required:"true" default:"3s"`
	}
)
