package repository

import "context"

type Transactional interface {
	OpenTx(ctx context.Context, fn func(c context.Context) error) error
}

type Order interface {
	Transactional
}
