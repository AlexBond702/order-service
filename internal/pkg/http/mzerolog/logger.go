package mzerolog

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/AlexBond702/order-service/internal/pkg/http/httph"
)

type (
	CallBackExtractorString = func(r *http.Request) string
	CallBackExtractorAny    = func(r *http.Request) any
	ExtractorStr            struct {
		key string
		ext CallBackExtractorString
	}
)

type ExtractorAny struct {
	key string
	ext CallBackExtractorAny
}
type middleware struct {
	log zerolog.Logger

	fromOption struct {
		skipper         func(r *http.Request) bool
		extStrOnSuccess []ExtractorStr
		extAnyOnSuccess []ExtractorAny
		extStrOnFail    []ExtractorStr
		extAnyOnFail    []ExtractorAny
	}
}

func (m *middleware) applyExtractors(
	r *http.Request,
	ev *zerolog.Event,
	extractorsStr []ExtractorStr,
	extractorsAny []ExtractorAny,
) {
	for i, n := 0, len(extractorsStr); i < n; i++ {
		str := extractorsStr[i]
		valueStr := str.ext(r)
		if valueStr != "" {
			ev.Str(str.key, valueStr)
		}
	}
	for i, n := 0, len(extractorsAny); i < n; i++ {
		an := extractorsAny[i]
		valueAny := an.ext(r)
		if valueAny != nil {
			ev.Any(an.key, valueAny)
		}
	}
}

func (m *middleware) Callback(c *gin.Context) {
	const (
		TailSuccess = " finished with no error"
		TailFail    = " finished (or aborted) with error"
	)
	start := time.Now()
	c.Next()
	r := c.Request
	err := httph.ErrorGet(r)
	execTime := time.Since(start)

	if m.fromOption.skipper(r) {
		return
	}

	var mb strings.Builder
	mb.Grow(48 + len(r.RequestURI))
	mb.WriteString(r.Method)
	mb.WriteByte(' ')
	mb.WriteString(r.RequestURI)

	var ev *zerolog.Event
	var extString []ExtractorStr
	var extAny []ExtractorAny

	if err == nil {
		mb.WriteString(TailSuccess)
		ev = m.log.Debug()
		extString = m.fromOption.extStrOnSuccess
		extAny = m.fromOption.extAnyOnSuccess
	} else {
		mb.WriteString(TailFail)
		ev = m.log.Error()
		extString = m.fromOption.extStrOnFail
		extAny = m.fromOption.extAnyOnFail
	}
	m.applyExtractors(r, ev, extString, extAny)

	ev.Err(err).
		Str("exec_time", execTime.String()).
		Str("client_ip", r.RemoteAddr).
		Msg(mb.String())
}
