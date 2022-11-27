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
	GetRepositoryList(model.RepositoryListRequest) ([]model.RepositoryListResponse, serror.SError)
	AddRepository(model.AddRepositoryRequest) (model.AddRepositoryResponse, serror.SError)
	EditRepository(model.EditRepositoryRequest) (model.EditRepositoryResponse, serror.SError)
	DeleteRepository(int64) serror.SError
}

type IScanningUsecase interface {
	GetScanningList(model.ScanningListRequest) ([]model.ScanningListResponse, serror.SError)
	AddNewScanning(int64) (model.ScanningResponse, serror.SError)

	// Start scanning the queue
	StartScanningInQueue() (errx serror.SError)
}

type IGrabScanner interface {
	StartScanningSession(string) ([]byte, serror.SError)
}
