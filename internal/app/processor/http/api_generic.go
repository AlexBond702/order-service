package processor

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	rhandler "github.com/AlexBond702/order-service/internal/app/handler"
)

func GenericRegHealthCheck(r *gin.Engine, h rhandler.Health) {
	regRoute(r, http.MethodGet, "/health", h.LastCheck, "api.generic.health_check")
}

func GenericRegMetrics(r *gin.Engine) {
	regRoute(r, http.MethodGet, "/metrics", gin.WrapH(promhttp.Handler()), "api.generic.metrics")
}

func GenericRegPprof(r *gin.Engine) {
	r.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof", gin.WrapF(pprofIndex))
	r.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	r.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	r.POST("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	r.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	r.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))

	r.GET("/debug/pprof/allocs", gin.WrapH(pprof.Handler("allocs")))
	r.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))
	r.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	r.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))
	r.GET("/debug/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
	r.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}

func pprofIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
}
