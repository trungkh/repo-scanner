package usecase

import (
	"fmt"
	"net/http"
	"repo-scanner/internal"
	"repo-scanner/internal/constants"
	"repo-scanner/internal/model"
	"repo-scanner/internal/utils/serror"
	"time"

	"github.com/jmoiron/sqlx/types"
	log "github.com/sirupsen/logrus"
)

type scanningUsecase struct {
	repositoryRepository internal.IRepositoryRepository
	scanningRepository   internal.IScanningRepository
	trxRepository        internal.ITrxRepository
	grabScanner          internal.IGrabScanner
}

func NewScanningUsecase(store internal.RepositoryStore, trxRepo internal.ITrxRepository, grabScanner internal.IGrabScanner) internal.IScanningUsecase {
	return scanningUsecase{
		repositoryRepository: store.RepositoryRepo,
		scanningRepository:   store.ScanningRepo,
		trxRepository:        trxRepo,
		grabScanner:          grabScanner,
	}
}

func (s scanningUsecase) GetScanningList(req model.ScanningListRequest) (res []model.ScanningListResponse, errx serror.SError) {
	res, errx = s.scanningRepository.GetScanningList(req)
	if errx != nil {
		errx.AddComments("[usecase][GetScanningList] while get scanning list")
		return
	}

	return
}

func (s scanningUsecase) AddNewScanning(repo_id int64) (res model.ScanningResponse, errx serror.SError) {
	var repo *model.Repository
	repo, errx = s.repositoryRepository.GetRepositoryById(repo_id)
	if errx != nil {
		errx.AddCommentf("[usecase][AddNewScanning] while GetRepositoryById (repository_id: %v)", repo_id)
		return
	} else if repo == nil {
		errx = serror.Newi(http.StatusBadRequest, "Repository not found|Repository not found")
		return
	} else if repo.IsActive == false {
		errx = serror.Newi(http.StatusBadRequest, "Repository is inactive|Repository is inactive")
		return
	}

	var tx *model.Trx
	tx, errx = s.trxRepository.Create()
	if errx != nil {
		errx.AddComments("[usecase][AddNewScanning] while create new transaction")
		return
	}
	defer func() {
		if errx != nil {
			errs := tx.Rollback()
			if errs != nil {
				log.Error("[usecase][AddNewScanning] Failed to rollback")
			}
		}
	}()

	res, errx = s.scanningRepository.AddNewScanning(tx, repo_id)
	if errx != nil {
		errx.AddComments("[usecase][AddNewScanning] while add repository")
		return
	}

	if errx == nil {
		err := tx.Commit()
		if err != nil {
			errx = serror.NewFromError(err)
			errx.AddCommentf("[usecase][AddNewScanning] Failed to commit transaction")
			return
		}
	}

	if constants.ScanningInProgress == false {
		go s.StartScanningInQueue()
	}
	return
}

func (s scanningUsecase) StartScanningInQueue() (errx serror.SError) {
	constants.ScanningInProgress = true
	log.Info("Start scanning...")
	for {
		// Limit should be 1 avoiding racing condition occuring in microservice architect
		req := model.ScanningListRequest{
			Limit:  1,
			Page:   1,
			Sort:   "asc",
			Status: "queued",
		}

		var scanningQueue []model.ScanningListResponse
		scanningQueue, errx = s.scanningRepository.GetScanningList(req)
		if errx != nil {
			errx.AddComments("[usecase][StartScanning] while get scanning list")
			return
		}
		if len(scanningQueue) == 0 {
			break
		}

		for idx := 0; idx < len(scanningQueue); idx++ {
			log.Infof("Scanning id[%v]...", scanningQueue[idx].Id)

			// Update status 'in_progress' immediately without creating a DB transaction
			_, errx = s.scanningRepository.EditScanningStatusById(nil,
				scanningQueue[idx].Id, constants.ScanningStatusInProgress, types.JSONText([]byte(`{}`)))
			if errx != nil {
				log.Error(errx)
				errx.AddComments("[usecase][StartScanning] while update scanning id[%v] status[%v]",
					fmt.Sprint(scanningQueue[idx].Id), constants.ScanningStatusInProgress)
				continue
			}
			log.Infof("Update scanning id[%v] status[%v] done",
				scanningQueue[idx].Id, constants.ScanningStatusInProgress)

			//Start scanning session by grab
			var res []byte
			var status string
			res, errx = s.grabScanner.StartScanningSession(scanningQueue[idx].Url)
			if errx != nil {
				log.Error(errx)
				status = constants.ScanningStatusFailure
			} else {
				status = constants.ScanningStatusSuccess
			}

			// Update status 'success/failure' immediately without creating a DB transaction
			_, errx = s.scanningRepository.EditScanningStatusById(nil,
				scanningQueue[idx].Id, status, types.JSONText(res))
			if errx != nil {
				log.Error(errx)
				errx.AddComments("[usecase][StartScanning] while update scanning id[%v] status[%v]",
					fmt.Sprint(scanningQueue[idx].Id), status)
				continue
			}
			log.Infof("Update scanning id[%v] status[%v] done", scanningQueue[idx].Id, status)
		}

		// Break time before checking any new ones from DB
		time.Sleep(1 * time.Second)
	}
	log.Info("Scanning done")
	constants.ScanningInProgress = false
	return
}
