package postgres

import (
	"database/sql"
	"repo-scanner/internal"
	"repo-scanner/internal/constants"
	"repo-scanner/internal/model"
	"repo-scanner/internal/repository/database"
	"repo-scanner/internal/repository/postgres/queries"
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/sqlq"
	"repo-scanner/internal/utils/uttime"
)

type repositoryRepository struct {
	psql
	Driver sqlq.SQLDriver
}

func NewRepositoryRepository(db *database.DB, q sqlq.SQLQuery, trxRepo internal.ITrxRepository) internal.IRepositoryRepository {
	return &repositoryRepository{
		psql: psql{
			//NumberHelp: numberHelp,
			TrxRepo: trxRepo,
			DB:      db.DB,
			Q:       q,
		},
		Driver: q.Driver(),
	}
}

func (r repositoryRepository) GetRepositoryList(req model.RepositoryListRequest) (res []model.RepositoryListResponse, errx serror.SError) {
	rows, err := r.DB.Queryx(queries.GetRepositoryList, req.Limit, (req.Page-1)*req.Limit)
	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[repository][GetRepositoryList] while get repository list")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r model.RepositoryListResponse
		if err = rows.StructScan(&r); err != nil {
			errx = serror.NewFromError(err)
			errx.AddCommentf("[repository][GetRepositoryList] while rows.StructScan")
			return
		}
		res = append(res, r)
	}
	return
}

func (r repositoryRepository) GetRepositoryById(repo_id int64) (res *model.Repository, errx serror.SError) {
	var repo model.Repository
	err := r.DB.QueryRowx(queries.GetRepositoryById, repo_id).StructScan(&repo)
	if err != nil {
		if err == sql.ErrNoRows {
			return
		}
		errx = serror.NewFromError(err)
		errx.AddCommentf("Get repository by repository_id query exec failed")
		return
	}

	return &repo, nil
}

func (r repositoryRepository) AddRepository(tx *model.Trx, req model.AddRepositoryRequest) (res model.AddRepositoryResponse, errx serror.SError) {
	currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)

	var err error
	if tx != nil {
		err = tx.QueryRowx(queries.InsertNewRepository,
			req.Name,
			req.Url,
			"Anonymous", // suppose someone else to add new repo
			currentTime,
			"Anonymous", // suppose someone else to modify repo
			currentTime,
		).StructScan(&res)
	} else {
		err = r.psql.DB.QueryRowx(queries.InsertNewRepository,
			req.Name,
			req.Url,
			"Anonymous", // suppose someone else to add new repo
			currentTime,
			"Anonymous", // suppose someone else to modify repo
			currentTime,
		).StructScan(&res)
	}

	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[repository][AddRepository] while add repository")
		return
	}
	return
}

func (r repositoryRepository) EditRepository(tx *model.Trx, req model.EditRepositoryRequest) (res model.EditRepositoryResponse, errx serror.SError) {
	currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)

	var err error
	if tx != nil {
		err = tx.QueryRowx(queries.EditRepository,
			req.Id,
			req.Name != nil, req.Name,
			req.Url != nil, req.Url,
			req.IsActive != nil, req.IsActive,
			"Anonymous", // suppose someone else to modify repo
			currentTime,
		).StructScan(&res)
	} else {
		err = r.psql.DB.QueryRowx(queries.EditRepository,
			req.Id,
			req.Name != nil, req.Name,
			req.Url != nil, req.Url,
			req.IsActive != nil, req.IsActive,
			"Anonymous", // suppose someone else to modify repo
			currentTime,
		).StructScan(&res)
	}

	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[repository][EditRepository] while add repository")
		return
	}
	return
}

func (r repositoryRepository) DeleteRepository(tx *model.Trx, repo_id int64) (errx serror.SError) {
	currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)

	var err error
	if tx != nil {
		_, err = tx.Exec(queries.DeleteRepository,
			repo_id,
			"Anonymous", // suppose someone else to delete repo
			currentTime)
	} else {
		_, err = r.psql.DB.Exec(queries.DeleteRepository,
			repo_id,
			"Anonymous", // suppose someone else to delete repo
			currentTime)
	}

	if err != nil {
		errx = serror.NewFromError(err)
		errx.AddCommentf("[repository][DeleteRepository] while delete repository")
		return
	}
	return
}
