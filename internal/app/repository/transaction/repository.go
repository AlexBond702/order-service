package transaction

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	repository "github.com/AlexBond702/order-service/internal/app/repository/conn/postgres"
)

type Repository struct {
	db *gorm.DB
}

func NewTxRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) OpenTx(ctx context.Context, fn func(c context.Context) error) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		c := repository.CtxWithTx(ctx, tx)
		return fn(c)
	}); err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	return nil
}
