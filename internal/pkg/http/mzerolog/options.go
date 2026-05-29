package mzerolog

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Option = func(m *middleware)

func WithLogger(l zerolog.Logger) Option {
	return func(m *middleware) {
		m.log = l
	}
}

func WithSkipper(skipper func(r *http.Request) bool) Option {
	if skipper != nil {
		return func(m *middleware) {
			m.fromOption.skipper = skipper
		}
	}
	return nil
}

func NewMiddleware(opts ...Option) gin.HandlerFunc {
	m := middleware{
		log: log.Logger,
	}
	m.fromOption.skipper = defaultSkipper
	for _, opt := range opts {
		opt(&m)
	}
	return m.Callback
}

func defaultSkipper(_ *http.Request) bool {
	return false
}

func WithStringExtractor(key string, callback CallBackExtractorString) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := ExtractorStr{key, callback}
			m.fromOption.extStrOnSuccess = append(m.fromOption.extStrOnSuccess, ext)
			m.fromOption.extStrOnFail = append(m.fromOption.extStrOnFail, ext)
		}
	}
}

func WithStringExtractorOnSuccess(key string, callback CallBackExtractorString) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := ExtractorStr{key, callback}
			m.fromOption.extStrOnSuccess = append(m.fromOption.extStrOnSuccess, ext)
		}
	}
}

func WithStringExtractorOnFail(key string, callback CallBackExtractorString) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := ExtractorStr{key, callback}
			m.fromOption.extStrOnFail = append(m.fromOption.extStrOnFail, ext)
		}
	}
}

func WithAnyExtractor(key string, callback CallBackExtractorAny) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := ExtractorAny{key, callback}
			m.fromOption.extAnyOnSuccess = append(m.fromOption.extAnyOnSuccess, ext)
			m.fromOption.extAnyOnFail = append(m.fromOption.extAnyOnFail, ext)
		}
	}
}

func WithAnyExtractorOnSuccess(key string, callback CallBackExtractorAny) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := ExtractorAny{key, callback}
			m.fromOption.extAnyOnSuccess = append(m.fromOption.extAnyOnSuccess, ext)
		}
	}
}

func WithAnyExtractorOnFail(key string, callback CallBackExtractorAny) Option {
	return func(m *middleware) {
		if key != "" && callback != nil {
			ext := ExtractorAny{key, callback}
			m.fromOption.extAnyOnFail = append(m.fromOption.extAnyOnFail, ext)
		}
	}
}
