package model

import (
	"github.com/jmoiron/sqlx"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utarray"
	"repo-scanner/internal/utils/utfunc"

	log "github.com/sirupsen/logrus"
)

type (
	Trx struct {
		*sqlx.Tx
		DB             *sqlx.DB
		parent         *Trx
		abortCallbacks []TrxCallbackFN
		admitCallbacks []TrxCallbackFN
		status         TrxStatus
	}

	TrxStatus     string
	TrxCallbackFN func()
)

const (
	TrxStatusActive   TrxStatus = "active"
	TrxStatusAdmitted TrxStatus = "admitted"
	TrxStatusAborted  TrxStatus = "aborted"
)

func (ox *Trx) Status() TrxStatus {
	if ox.IsActive() {
		ox.status = TrxStatusActive
	}

	if !ox.IsActive() && !utarray.IsExist(ox.status, []TrxStatus{
		TrxStatusAborted,
		TrxStatusAdmitted,
	}) {
		ox.status = TrxStatusAborted
		if ox.parent != nil {
			ox.status = ox.parent.Status()
		}
	}

	return ox.status
}

func (ox *Trx) IsActive() bool {
	return ((ox.parent != nil && ox.parent.Tx != nil) || ox.parent == nil) && ox.Tx != nil
}

func (ox *Trx) Admit() (errx serror.SError) {
	if !ox.IsActive() {
		return errx
	}

	err := ox.Tx.Commit()
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to committing")

		errs := ox.Abort()
		if errs != nil {
			errs.AddComments("while abort")
			log.Warn(errs)
		}
	}

	if errx == nil {
		ox.Tx = nil
		ox.status = TrxStatusAdmitted

		if len(ox.admitCallbacks) > 0 {
			for _, v := range ox.admitCallbacks {
				errs := utfunc.Try(func() serror.SError {
					v()
					return nil
				})
				if errs != nil {
					errs.AddComments("while call admit callback")
					log.Error(errs)
				}
			}
		}
	}
	return errx
}

func (ox *Trx) Abort() (errx serror.SError) {
	if !ox.IsActive() {
		return errx
	}

	err := ox.Tx.Rollback()
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to rollback")
	}

	if errx == nil {
		ox.Tx = nil
		ox.status = TrxStatusAborted

		if len(ox.abortCallbacks) > 0 {
			for _, v := range ox.abortCallbacks {
				errs := utfunc.Try(func() serror.SError {
					v()
					return nil
				})
				if errs != nil {
					errs.AddComments("while call abort callback")
					log.Error(errs)
				}
			}
		}
	}

	return errx
}

func (ox *Trx) refresh() (errx serror.SError) {
	var err error
	ox.Tx, err = ox.DB.Beginx()
	if err != nil {
		errx = serror.NewFromErrorc(err, "Failed to begin transaction")
		return errx
	}

	return errx
}

func (ox *Trx) AbortRefresh(force bool) (errx serror.SError) {
	return serror.New("function still draft")

	// if ox.Tx != nil {
	// 	err := ox.Tx.Rollback()
	// 	if err != nil {
	// 		errs := serror.NewFromErrorc(err, "Failed to rollback")

	// 		if force {
	// 			logger.Err(errs)
	// 		}

	// 		if !force {
	// 			errx = errs
	// 			return errx
	// 		}
	// 	}
	// }

	// return ox.refresh()
}

func (ox *Trx) AdmitRefresh(force bool) (errx serror.SError) {
	return serror.New("function still draft")

	// if ox.Tx != nil {
	// 	err := ox.Tx.Commit()
	// 	if err != nil {
	// 		errs := serror.NewFromErrorc(err, "Failed to committing")

	// 		if force {
	// 			logger.Err(errs)
	// 		}

	// 		if !force {
	// 			errx = errs
	// 			return errx
	// 		}
	// 	}
	// }

	// return ox.refresh()
}

func (ox *Trx) AbortCallback(fn TrxCallbackFN) {
	if fn == nil {
		return
	}

	switch ox.Status() {
	case TrxStatusActive:
		ox.abortCallbacks = append(ox.abortCallbacks, fn)

	case TrxStatusAborted:
		errs := utfunc.Try(func() serror.SError {
			fn()
			return nil
		})
		if errs != nil {
			errs.AddComments("while call abort callback")
			log.Error(errs)
		}
	}
}

func (ox *Trx) AdmitCallback(fn TrxCallbackFN) {
	if fn == nil {
		return
	}

	switch ox.Status() {
	case TrxStatusActive:
		ox.admitCallbacks = append(ox.admitCallbacks, fn)

	case TrxStatusAdmitted:
		errs := utfunc.Try(func() serror.SError {
			fn()
			return nil
		})
		if errs != nil {
			errs.AddComments("while call admit callback")
			log.Error(errs)
		}
	}
}

func (ox *Trx) ForkTrx() (trx *Trx, errx serror.SError) {
	tx, err := ox.DB.Beginx()
	if err != nil {
		errx = serror.NewFromErrorc(err, "while begin transaction")
		return trx, errx
	}

	trx = &Trx{
		Tx: tx,
		DB: ox.DB,
	}

	// linkin parent
	switch {
	case ox.parent != nil:
		trx.parent = ox.parent

	default:
		trx.parent = ox
	}

	// hook when admit or abort from parent
	{
		trx.parent.AdmitCallback(func() {
			errs := trx.Admit()
			if errs != nil {
				errs.AddComments("while admit from parent")
				log.Error(errs)
			}
		})

		trx.parent.AbortCallback(func() {
			errs := trx.Abort()
			if errs != nil {
				errs.AddComments("while abort from parent")
				log.Error(errs)
			}
		})
	}

	return trx, errx
}
