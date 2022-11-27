package internal

import (
	"repo-scanner/internal/model"
	"repo-scanner/internal/utils/serror"

	"github.com/jmoiron/sqlx/types"
)

type INumberHelperRepository interface {
	Format(amount float64, precision int) (res string)
	FormatPlain(amount float64, precision int) (res string)
	Round(amount float64) float64
}

type ITrxRepository interface {
	Create() (trx *model.Trx, errx serror.SError)
}

type RepositoryStore struct {
	ScanningRepo   IScanningRepository
	RepositoryRepo IRepositoryRepository
}

type IRepositoryRepository interface {
	GetRepositoryList(model.RepositoryListRequest) ([]model.RepositoryListResponse, serror.SError)
	GetRepositoryById(repo_id int64) (*model.Repository, serror.SError)
	AddRepository(*model.Trx, model.AddRepositoryRequest) (model.AddRepositoryResponse, serror.SError)
	EditRepository(*model.Trx, model.EditRepositoryRequest) (model.EditRepositoryResponse, serror.SError)
	DeleteRepository(*model.Trx, int64) serror.SError
}

type IScanningRepository interface {
	GetScanningList(model.ScanningListRequest) ([]model.ScanningListResponse, serror.SError)
	AddNewScanning(*model.Trx, int64) (model.ScanningResponse, serror.SError)
	EditScanningStatusById(*model.Trx, int64, string, types.JSONText) (model.ScanningResponse, serror.SError)
}
