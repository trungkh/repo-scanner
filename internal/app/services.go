package config

import (
	"repo-scanner/internal"
	"repo-scanner/internal/utils/serror"

	"repo-scanner/internal/delivery/rest"
	"repo-scanner/internal/repository/postgres"
	"repo-scanner/internal/repository/scanner"
	"repo-scanner/internal/usecase"
)

func (c *Config) InitService() serror.SError {
	// trx
	trxRepo := postgres.NewTrxRepository(c.DB.DB)

	repositoryRepo := postgres.NewRepositoryRepository(c.DB, c.Query, trxRepo)
	scanningRepo := postgres.NewScanningRepository(c.DB, c.Query, trxRepo)
	repoStore := internal.RepositoryStore{
		RepositoryRepo: repositoryRepo,
		ScanningRepo:   scanningRepo,
	}

	grabScanner := scanner.NewGrabScanner(repoStore)

	repositoryUsecase := usecase.NewRepositoryUsecase(repoStore, trxRepo)
	scanningUsecase := usecase.NewScanningUsecase(repoStore, trxRepo, grabScanner)
	usecaseStore := internal.UsecaseStore{
		RepositoryUsecase: repositoryUsecase,
		ScanningUsecase:   scanningUsecase,
	}

	rest.NewHandler(c.Server, usecaseStore)

	return nil
}
