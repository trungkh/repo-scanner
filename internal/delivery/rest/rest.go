package rest

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"

	"repo-scanner/internal"
	"repo-scanner/internal/constants"
	"repo-scanner/internal/model"
	"repo-scanner/internal/utils/response"
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utint"
)

type handler struct {
	repositoryUseCase internal.IRepositoryUsecase
	scanningUsecase   internal.IScanningUsecase
}

func (hd handler) GetRepositoryList(ctx *gin.Context) {
	var (
		errx serror.SError
	)

	defer func() {
		if errx != nil {
			log.Error(errx.Comments())
		}
	}()

	log.Infof("GetRepositoryList invoked")

	req := model.RepositoryListRequest{
		Limit: utint.StringToInt(ctx.Query("limit"), constants.DefaultLimit),
		Page:  utint.StringToInt(ctx.Query("page"), constants.DefaultPage),
	}

	err := validator.New().Struct(req)
	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddComments("[delivery][GetRepositoryList] while validate struct")
		response.ResultError(ctx, response.ErrorQueryValidationFail, err)
		return
	}

	var res []model.RepositoryListResponse
	res, errx = hd.repositoryUseCase.GetRepositoryList(req)
	if errx != nil {
		errx.AddCommentf("[delivery][GetRepositoryList] while get repository list")
		if errx.Code() < 1 {
			errx = serror.Newic(http.StatusInternalServerError, errx.Error(), errx.Comments())
		}
		response.ResultSError(ctx, errx)
		return
	}

	response.ResultWithData(ctx, response.SuccessGetDataOk, res)
	return
}

func (hd handler) AddRepository(ctx *gin.Context) {
	var (
		errx serror.SError
	)

	defer func() {
		if errx != nil {
			log.Error(errx.Comments())
		}
	}()

	log.Infof("AddRepository invoked")

	req := model.AddRepositoryRequest{}
	ctx.BindJSON(&req)

	err := validator.New().Struct(req)
	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[delivery][AddRepository] while validate struct")
		response.ResultError(ctx, response.ErrorPayloadValidationFail, err)
		return
	}

	pathParts := strings.Split(req.Url, "/")
	idx := 0
	for ; idx < len(pathParts); idx++ {
		if (pathParts[idx] == "github.com" ||
			pathParts[idx] == "gitlab.com" ||
			pathParts[idx] == "bitbucket.org") && len(pathParts[idx:]) == 3 {
			req.Url = strings.Join(pathParts[idx:], "/")
			break
		}
	}
	if idx >= len(pathParts) {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[delivery][AddRepository] while validate url")
		response.ResultError(ctx, response.ErrorUrlValidationFail, err)
		return
	}

	var res model.AddRepositoryResponse
	res, errx = hd.repositoryUseCase.AddRepository(req)
	if errx != nil {
		errx.AddCommentf("[delivery][AddRepository] while add new repository")
		if errx.Code() < 1 {
			errx = serror.Newic(http.StatusInternalServerError, errx.Error(), errx.Comments())
		}
		response.ResultSError(ctx, errx)
		return
	}

	response.ResultWithData(ctx, response.SuccessCreated, res)
	return
}

func (hd handler) EditRepository(ctx *gin.Context) {
	var (
		errx serror.SError
	)

	defer func() {
		if errx != nil {
			log.Error(errx.Comments())
		}
	}()

	log.Infof("EditRepository invoked")

	req := model.EditRepositoryRequest{
		Id: utint.StringToInt(ctx.Param("repository_id"), 0),
	}
	ctx.BindJSON(&req)

	if req.Id <= 0 {
		errx = serror.New("Invalid repository_id")
		response.ResultError(ctx, response.ErrorParamValidationFail, errx)
		return
	}

	err := validator.New().Struct(req)
	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[delivery][EditRepository] while validate struct")
		response.ResultError(ctx, response.ErrorPayloadValidationFail, err)
		return
	}

	if req.Url != nil {
		pathParts := strings.Split(*req.Url, "/")
		idx := 0
		for ; idx < len(pathParts); idx++ {
			if (pathParts[idx] == "github.com" ||
				pathParts[idx] == "gitlab.com" ||
				pathParts[idx] == "bitbucket.org") && len(pathParts[idx:]) == 3 {
				*req.Url = strings.Join(pathParts[idx:], "/")
				break
			}
		}
		if idx >= len(pathParts) {
			errx = serror.NewFromError(err)
			errx.AddCommentf("[delivery][AddRepository] while validate url")
			response.ResultError(ctx, response.ErrorUrlValidationFail, err)
			return
		}
	}

	var res model.EditRepositoryResponse
	res, errx = hd.repositoryUseCase.EditRepository(req)
	if errx != nil {
		errx.AddCommentf("[delivery][EditRepository] while edit repository")
		if errx.Code() < 1 {
			errx = serror.Newic(http.StatusInternalServerError, errx.Error(), errx.Comments())
		}
		response.ResultSError(ctx, errx)
		return
	}

	response.ResultWithData(ctx, response.SuccessUpdated, res)
	return
}

func (hd handler) DeleteRepository(ctx *gin.Context) {
	var (
		errx serror.SError
	)

	defer func() {
		if errx != nil {
			log.Error(errx.Comments())
		}
	}()

	log.Infof("DeleteRepository invoked")

	repoId := utint.StringToInt(ctx.Param("repository_id"), 0)
	if repoId <= 0 {
		errx = serror.New("Invalid repository_id")
		response.ResultError(ctx, response.ErrorParamValidationFail, errx)
		return
	}

	errx = hd.repositoryUseCase.DeleteRepository(repoId)
	if errx != nil {
		errx.AddCommentf("[delivery][EditRepository] while edit repository")
		if errx.Code() < 1 {
			errx = serror.Newic(http.StatusInternalServerError, errx.Error(), errx.Comments())
		}
		response.ResultSError(ctx, errx)
		return
	}

	response.Result(ctx, response.SuccessDeleted)
	return
}

func (hd handler) TriggerRepoScanning(ctx *gin.Context) {
	var (
		errx serror.SError
	)

	defer func() {
		if errx != nil {
			log.Error(errx.Comments())
		}
	}()

	log.Infof("TriggerRepoScanning invoked")

	repo_id := utint.StringToInt(ctx.Param("repository_id"), 0)
	if repo_id <= 0 {
		errx = serror.New("Invalid repository_id")
		response.ResultError(ctx, response.ErrorParamValidationFail, errx)
		return
	}

	var res model.ScanningResponse
	res, errx = hd.scanningUsecase.AddNewScanning(repo_id)
	if errx != nil {
		errx.AddCommentf("[delivery][TriggerRepoScanning] while add new scanning")
		if errx.Code() < 1 {
			errx = serror.Newic(http.StatusInternalServerError, errx.Error(), errx.Comments())
		}
		response.ResultSError(ctx, errx)
		return
	}

	response.ResultWithData(ctx, response.SuccessCreated, res)
	return
}

func (hd handler) ScanningResult(ctx *gin.Context) {
	var (
		errx serror.SError
	)

	defer func() {
		if errx != nil {
			log.Error(errx.Comments())
		}
	}()

	log.Infof("ScanningResult invoked")

	req := model.ScanningListRequest{
		Limit: utint.StringToInt(ctx.Query("limit"), constants.DefaultLimit),
		Page:  utint.StringToInt(ctx.Query("page"), constants.DefaultPage),
	}

	var ok bool
	if req.Sort, ok = ctx.GetQuery("sort"); ok == false {
		req.Sort = "desc"
	}
	if req.Status, ok = ctx.GetQuery("status"); ok == false {
		req.Status = "all"
	}

	err := validator.New().Struct(req)
	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[delivery][ScanningResult] while validate struct")
		response.ResultError(ctx, response.ErrorQueryValidationFail, err)
		return
	}

	var res []model.ScanningListResponse
	res, errx = hd.scanningUsecase.GetScanningList(req)
	if errx != nil {
		errx.AddCommentf("[delivery][ScanningResult] while get scanning list")
		if errx.Code() < 1 {
			errx = serror.Newic(http.StatusInternalServerError, errx.Error(), errx.Comments())
		}
		response.ResultSError(ctx, errx)
		return
	}

	response.ResultWithData(ctx, response.SuccessGetDataOk, res)
	return
}
