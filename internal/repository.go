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

type ITrx interface {
	Admit() serror.SError
	Abort() serror.SError
}

type ITrxRepository interface {
	/* Create database transaction */
	Create() (*model.Trx, serror.SError)
}

type RepositoryStore struct {
	ScanningRepo   IScanningRepository
	RepositoryRepo IRepositoryRepository
}

type IRepositoryRepository interface {
	// Get list of git repositories
	GetRepositoryList(model.RepositoryListRequest) ([]model.RepositoryListResponse, serror.SError)

	// Get git repository detail by given repository id
	GetRepositoryById(repo_id int64) (*model.Repository, serror.SError)

	// Insert new repository by given name and url
	AddRepository(*model.Trx, model.AddRepositoryRequest) (model.AddRepositoryResponse, serror.SError)

	// Update existing repository by given repository id.
	// Required: repository Id.
	// Optional: Name, Url and IsActive.
	EditRepository(*model.Trx, model.EditRepositoryRequest) (model.EditRepositoryResponse, serror.SError)

	// Delete existing repository by given repository id
	DeleteRepository(*model.Trx, int64) serror.SError
}

type IScanningRepository interface {
	// Get list of recently scanned
	GetScanningList(model.ScanningListRequest) ([]model.ScanningListResponse, serror.SError)

	// Insert new scanning by given active repository id
	AddNewScanning(*model.Trx, int64) (model.ScanningResponse, serror.SError)

	// Update status of existing scanning by given repository id
	// Required: repository Id, status
	// Optional: Findings
	EditScanningStatusById(*model.Trx, int64, string, types.JSONText) (model.ScanningResponse, serror.SError)
}
