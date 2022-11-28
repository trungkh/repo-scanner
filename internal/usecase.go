package internal

import (
	"repo-scanner/internal/model"
	"repo-scanner/internal/utils/serror"
)

type UsecaseStore struct {
	RepositoryUsecase IRepositoryUsecase
	ScanningUsecase   IScanningUsecase
}

type IRepositoryUsecase interface {
	// Get list of git repositories
	GetRepositoryList(model.RepositoryListRequest) ([]model.RepositoryListResponse, serror.SError)

	// Create new repository by given name and url
	AddRepository(model.AddRepositoryRequest) (model.AddRepositoryResponse, serror.SError)

	// Edit existing repository by given repository id.
	// Required: repository Id.
	// Optional: Name, Url and IsActive.
	EditRepository(model.EditRepositoryRequest) (model.EditRepositoryResponse, serror.SError)

	// Delete existing repository by given repository id
	DeleteRepository(int64) serror.SError
}

type IScanningUsecase interface {
	// Get list of recently scanned
	GetScanningList(model.ScanningListRequest) ([]model.ScanningListResponse, serror.SError)

	// Create new scanning by given active repository id
	AddNewScanning(int64) (model.ScanningResponse, serror.SError)

	// Start scanning from queue
	StartScanningInQueue() (errx serror.SError)
}

type IGrabScanner interface {
	// Start scanning session with git repository url
	StartScanningSession(string) ([]byte, serror.SError)
}
