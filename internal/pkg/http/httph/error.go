package httph

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

type Error struct {
	Message string `json:"error"`
}

type _ContextKeyError struct{}

type _ContextValueError struct {
	err       error
	detail    string
	isHandled bool
}
type Middleware = gin.HandlerFunc

func errorPrepare(ctx context.Context) context.Context {
	errCtx := new(_ContextValueError)
	return context.WithValue(ctx, _ContextKeyError{}, errCtx)
}

func errorGet(ctx context.Context) error {
	ctxValue := ctx.Value(_ContextKeyError{})
	if ctxValue != nil {
		if errV, ok := ctxValue.(*_ContextValueError); ok {
			return errV.err
		}
	}
	return nil
}

func errorGetDetail(ctx context.Context) string {
	ctxDetail := ctx.Value(_ContextKeyError{})
	if ctxDetail != nil {
		if errV, ok := ctxDetail.(*_ContextValueError); ok {
			return errV.detail
		}
	}
	return ""
}

func errorApply(ctx context.Context, err error) {
	ctxValue := ctx.Value(_ContextKeyError{})
	if ctxValue != nil {
		if errV, ok := ctxValue.(*_ContextValueError); ok {
			errV.err = err
			return
		}
	}
	if err != nil {
		trace.SpanFromContext(ctx).RecordError(err)
	}
}

func errorApplyDetail(ctx context.Context, detail string) {
	ctxValue := ctx.Value(_ContextKeyError{})
	if ctxValue != nil {
		if errV, ok := ctxValue.(*_ContextValueError); ok {
			errV.detail = detail
		}
	}
}

func errorTryAcquireHandling(ctx context.Context) bool {
	ctxValue := ctx.Value(_ContextKeyError{})
	if ctxValue != nil {
		if errV, ok := ctxValue.(*_ContextValueError); ok {
			if errV == nil || errV.isHandled {
				return false
			}
			errV.isHandled = true
		}
	}
	return true
}

func ErrorPrepare(r *http.Request) *http.Request {
	return r.WithContext(errorPrepare(r.Context()))
}

func ErrorGet(r *http.Request) error {
	ctx := r.Context()
	if err := errorGet(ctx); err != nil {
		return err
	}
	return nil
}

func ErrorGetDetail(r *http.Request) string {
	ctx := r.Context()
	if detail := errorGetDetail(ctx); detail != "" {
		return detail
	}
	return ""
}

func ErrorApply(r *http.Request, err error) {
	ctx := r.Context()
	errorApply(ctx, err)
}

func ErrorApplyDetail(r *http.Request, detail string) {
	ctx := r.Context()
	errorApplyDetail(ctx, detail)
}

func ErrorTryAcquireHandling(r *http.Request) bool {
	ctx := r.Context()
	return errorTryAcquireHandling(ctx)
}

func NewErrorMiddleware() Middleware {
	return func(c *gin.Context) {
		c.Request = ErrorPrepare(c.Request)

		c.Next()
	}
}
