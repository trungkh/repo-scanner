package postgres

import (
	"github.com/jmoiron/sqlx"

	"repo-scanner/internal"
	"repo-scanner/internal/model"
	"repo-scanner/internal/utils/serror"
)

type trx struct {
	DB *sqlx.DB
}

func NewTrxRepository(db *sqlx.DB) internal.ITrxRepository {
	return trx{
		DB: db,
	}
}

func (ox trx) Create() (trx *model.Trx, errx serror.SError) {
	tx, err := ox.DB.Beginx()
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to create transaction")
		return trx, errx
	}

	trx = &model.Trx{
		DB: ox.DB,
		Tx: tx,
	}
	return trx, errx
}
