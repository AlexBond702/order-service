package rcpostgres

import (
	"context"

	"gorm.io/gorm"
)

type _ctxKeyTx struct{}

func GetTxFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	value, ok := ctx.Value(_ctxKeyTx{}).(gorm.DB)
	if !ok {
		return db
	}
	return &value
}

func CtxWithTx(ctx context.Context, tx *gorm.DB) context.Context {
	ctxValueTx := context.WithValue(ctx, _ctxKeyTx{}, tx)
	return ctxValueTx
}
