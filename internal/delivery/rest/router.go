package rest

import (
	"repo-scanner/internal"

	"github.com/gin-gonic/gin"
)

func NewHandler(router *gin.Engine, store internal.UsecaseStore) {
	h := handler{
		repositoryUseCase: store.RepositoryUsecase,
		scanningUsecase:   store.ScanningUsecase,
	}

	// Repository handlers
	router.GET("/v1/repositories", h.GetRepositoryList)
	router.POST("/v1/repository", h.AddRepository)
	router.PUT("/v1/repository/:repository_id", h.EditRepository)
	router.DELETE("/v1/repository/:repository_id", h.DeleteRepository)
	router.POST("/v1/repository/:repository_id/scan", h.TriggerRepoScanning)

	// Scanning handlers
	router.GET("/v1/scanning/result", h.ScanningResult)

	// Start scanning immediately which are unfinished
	// Especially, adapting multiple services running simutanously
	go h.scanningUsecase.StartScanningInQueue()
}
