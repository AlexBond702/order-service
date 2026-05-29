package respondent

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AlexBond702/order-service/internal/pkg/http/httph"
)

type respondent struct {
	expander   Expander
	applicator Applicator
}

type HttpContext struct {
	W http.ResponseWriter
	R *http.Request
}

var ErrBadExpander = errors.New("respondent: expander is required")

func (rs *respondent) Callback(ctx any, err error) {
	manifest := rs.expander.Expand(err)
	if manifest == nil {
		return
	}
	rs.applicator.Apply(ctx, manifest)
}

func (rs *respondent) CallbackForHttp(w http.ResponseWriter, r *http.Request, err error) {
	ctxHttp := &HttpContext{
		W: w,
		R: r,
	}
	rs.Callback(ctxHttp, err)
}

func newRespondent(expander Expander, applicator Applicator) *respondent {
	if expander == nil {
		panic(ErrBadExpander)
	}
	if applicator == nil {
		applicator = NewSimpleApplicator()
	}
	return &respondent{
		expander:   expander,
		applicator: applicator,
	}
}

func NewMiddleware(expander Expander, applicator Applicator) gin.HandlerFunc {
	resp := newRespondent(expander, applicator)
	return func(c *gin.Context) {
		c.Next()
		err := httph.ErrorGet(c.Request)
		mayHandle := httph.ErrorTryAcquireHandling(c.Request)
		if err != nil && mayHandle {
			resp.CallbackForHttp(c.Writer, c.Request, err)
		}
	}
}
