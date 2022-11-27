package postgres

import (
	"fmt"
	"repo-scanner/internal"
	"repo-scanner/internal/constants"
	"repo-scanner/internal/model"
	"repo-scanner/internal/repository/database"
	"repo-scanner/internal/repository/postgres/queries"
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/sqlq"
	"repo-scanner/internal/utils/uttime"

	"github.com/jmoiron/sqlx/types"
)

type scanningRepository struct {
	psql
	Driver sqlq.SQLDriver
}

func NewScanningRepository(db *database.DB, q sqlq.SQLQuery, trxRepo internal.ITrxRepository) internal.IScanningRepository {
	return &scanningRepository{
		psql: psql{
			//NumberHelp: numberHelp,
			TrxRepo: trxRepo,
			DB:      db.DB,
			Q:       q,
		},
		Driver: q.Driver(),
	}
}

func (s scanningRepository) GetScanningList(req model.ScanningListRequest) (res []model.ScanningListResponse, errx serror.SError) {
	query := fmt.Sprintf(queries.GetScanningList, req.Sort)
	rows, err := s.DB.Queryx(query,
		req.Status,
		req.Limit,
		(req.Page-1)*req.Limit)
	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[repository][ScanningResult] while get repository list")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r model.ScanningListResponse
		if err = rows.StructScan(&r); err != nil {
			errx = serror.NewFromError(err)
			errx.AddCommentf("[repository][ScanningResult] while rows.StructScan")
			return
		}
		res = append(res, r)
	}
	return
}

func (s scanningRepository) AddNewScanning(tx *model.Trx, repo_id int64) (res model.ScanningResponse, errx serror.SError) {
	currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)

	var err error
	if tx != nil {
		err = tx.QueryRowx(queries.InsertNewScanning,
			repo_id,
			currentTime, // queued date
			"Anonymous", // suppose someone else to trigger scanning
			currentTime,
			"Anonymous", // suppose someone else to modify scanning
			currentTime,
		).StructScan(&res)
	} else {
		err = s.psql.DB.QueryRowx(queries.InsertNewScanning,
			repo_id,
			currentTime, // queued date
			"Anonymous", // suppose someone else to trigger scanning
			currentTime,
			"Anonymous", // suppose someone else to modify scanning
			currentTime,
		).StructScan(&res)
	}

	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[repository][AddNewScanning] while add new scanning")
		return
	}
	return
}

func (s scanningRepository) EditScanningStatusById(tx *model.Trx,
	scanningId int64, scanningStatus string, findings types.JSONText) (res model.ScanningResponse, errx serror.SError) {
	currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)

	var err error
	if tx != nil {
		switch scanningStatus {
		case constants.ScanningStatusInProgress:
			err = tx.QueryRowx(queries.UpdateScanningInProgressById,
				scanningId,
				scanningStatus,
				"Automated", // suppose someone else to modify repo
				currentTime,
			).StructScan(&res)

		case constants.ScanningStatusSuccess, constants.ScanningStatusFailure:
			err = tx.QueryRowx(queries.UpdateScanningFinishedById,
				scanningId,
				scanningStatus,
				findings,
				"Automated", // suppose someone else to modify repo
				currentTime,
			).StructScan(&res)
		}
	} else {
		switch scanningStatus {
		case constants.ScanningStatusInProgress:
			err = s.psql.DB.QueryRowx(queries.UpdateScanningInProgressById,
				scanningId,
				scanningStatus,
				"Automated", // suppose someone else to modify repo
				currentTime,
			).StructScan(&res)

		case constants.ScanningStatusSuccess, constants.ScanningStatusFailure:
			err = s.psql.DB.QueryRowx(queries.UpdateScanningFinishedById,
				scanningId,
				scanningStatus,
				findings,
				"Automated", // suppose someone else to modify repo
				currentTime,
			).StructScan(&res)
		}
	}

	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[repository][EditScanningStatusById] while update scanning status")
		return
	}
	return
}
