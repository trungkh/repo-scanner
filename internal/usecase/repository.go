package usecase

import (
	"net/http"
	"repo-scanner/internal"
	"repo-scanner/internal/model"
	"repo-scanner/internal/utils/serror"

	log "github.com/sirupsen/logrus"
)

type repositoryUsecase struct {
	repositoryRepository internal.IRepositoryRepository
	trxRepository        internal.ITrxRepository
}

func NewRepositoryUsecase(store internal.RepositoryStore, trxRepo internal.ITrxRepository) internal.IRepositoryUsecase {
	return repositoryUsecase{
		repositoryRepository: store.RepositoryRepo,
		trxRepository:        trxRepo,
	}
}

func (r repositoryUsecase) GetRepositoryList(req model.RepositoryListRequest) (res []model.RepositoryListResponse, errx serror.SError) {
	res, errx = r.repositoryRepository.GetRepositoryList(req)
	if errx != nil {
		errx.AddComments("[usecase][GetRepositoryList] while get repository list")
		return
	}
	return
}

func (r repositoryUsecase) AddRepository(req model.AddRepositoryRequest) (res model.AddRepositoryResponse, errx serror.SError) {
	var tx internal.ITrx
	tx, errx = r.trxRepository.Create()
	if errx != nil {
		errx.AddComments("[usecase][AddRepository] while create new transaction")
		return
	}
	defer func() {
		if errx != nil {
			errs := tx.Abort()
			if errs != nil {
				log.Error("[usecase][AddRepository] Failed to rollback")
			}
		}
	}()

	res, errx = r.repositoryRepository.AddRepository(tx.(*model.Trx), req)
	if errx != nil {
		errx.AddComments("[usecase][AddRepository] while add repository")
		return
	}

	if errx == nil {
		err := tx.Admit()
		if err != nil {
			errx = serror.NewFromError(err)
			errx.AddCommentf("[usecase][AddRepository] Failed to commit transaction")
			return
		}
	}
	return
}

func (r repositoryUsecase) EditRepository(req model.EditRepositoryRequest) (res model.EditRepositoryResponse, errx serror.SError) {
	if req.Name == nil &&
		req.Url == nil &&
		req.IsActive == nil {
		errx = serror.Newi(http.StatusNotAcceptable, "Nothing to update|Nothing to update")
		return
	}

	var originalRepo *model.Repository
	originalRepo, errx = r.repositoryRepository.GetRepositoryById(req.Id)
	if errx != nil {
		errx.AddCommentf("[usecase][EditRepository] while GetRepositoryById (repository_id: %v)", req.Id)
		return
	} else if originalRepo == nil {
		errx = serror.Newi(http.StatusBadRequest, "Repository not found|Repository not found")
		return
	}

	var is_duplicate = true
	switch {
	case req.Name != nil && originalRepo.Name != *req.Name:
		is_duplicate = false
	case req.Url != nil && originalRepo.Url != *req.Url:
		is_duplicate = false
	case req.IsActive != nil && originalRepo.IsActive != *req.IsActive:
		is_duplicate = false
	}
	if is_duplicate {
		errx = serror.Newi(http.StatusNotAcceptable, "Nothing to update|Nothing to update")
		return
	}

	var tx *model.Trx
	tx, errx = r.trxRepository.Create()
	if errx != nil {
		errx.AddComments("[usecase][EditRepository] while create new transaction")
		return
	}
	defer func() {
		if errx != nil {
			errs := tx.Abort()
			if errs != nil {
				log.Error("[usecase][EditRepository] Failed to rollback")
			}
		}
	}()

	res, errx = r.repositoryRepository.EditRepository(tx, req)
	if errx != nil {
		errx.AddCommentf("[usecase][EditRepository] while EditRepository (repository_id: %v)", req.Id)
		return
	}
	if errx == nil {
		err := tx.Admit()
		if err != nil {
			errx = serror.NewFromError(err)
			errx.AddCommentf("[usecase][EditRepository] Failed to commit transaction")
			return
		}
	}
	return
}

func (r repositoryUsecase) DeleteRepository(repo_id int64) (errx serror.SError) {
	var repo *model.Repository
	repo, errx = r.repositoryRepository.GetRepositoryById(repo_id)
	if errx != nil {
		errx.AddCommentf("[usecase][DeleteRepository] while GetRepositoryById (repository_id: %v)", repo_id)
		return
	} else if repo == nil {
		errx = serror.Newi(http.StatusBadRequest, "Repository not found|Repository not found")
		return
	}

	var tx *model.Trx
	tx, errx = r.trxRepository.Create()
	if errx != nil {
		errx.AddComments("[usecase][DeleteRepository] while create new transaction")
		return
	}
	defer func() {
		if errx != nil {
			errs := tx.Abort()
			if errs != nil {
				log.Error("[usecase][DeleteRepository] Failed to rollback")
			}
		}
	}()

	errx = r.repositoryRepository.DeleteRepository(tx, repo_id)
	if errx != nil {
		errx.AddCommentf("[usecase][DeleteRepository] while DeleteRepository (repository_id: %v)", repo_id)
		return
	}
	if errx == nil {
		err := tx.Admit()
		if err != nil {
			errx = serror.NewFromError(err)
			errx.AddCommentf("[usecase][DeleteRepository] Failed to commit transaction")
			return
		}
	}
	return
}
