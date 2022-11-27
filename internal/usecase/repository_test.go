package usecase

import (
	"errors"
	"testing"

	"repo-scanner/internal/mocks"
	"repo-scanner/internal/model"
	"repo-scanner/internal/utils/serror"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRepositoryList(t *testing.T) {
	repoMock := new(mocks.IRepositoryRepository)

	listTests := []struct {
		name    string
		mock    func()
		args    model.RepositoryListRequest
		want    interface{}
		wantErr bool
	}{
		{
			name: "ok",
			mock: func() {
				w := []model.RepositoryListResponse{
					{
						Id:       3,
						Name:     "JQuery",
						Url:      "github.com/jquery/jquery",
						IsActive: true,
					},
				}

				repoMock.On("GetRepositoryList", mock.Anything).Return(w, nil).Once()
			},
			args: model.RepositoryListRequest{
				Limit: 1,
				Page:  1,
			},
			want: []model.RepositoryListResponse{
				{
					Id:       3,
					Name:     "JQuery",
					Url:      "github.com/jquery/jquery",
					IsActive: true,
				},
			},
			wantErr: false,
		},
		{
			name: "error",
			mock: func() {
				e := errors.New("error")
				w := serror.NewFromErrorc(e, "[usecase][GetRepositoryList] get repository list")

				repoMock.On("GetRepositoryList", mock.Anything).
					Return([]model.RepositoryListResponse{}, w).Once()
			},
			want:    []model.RepositoryListResponse{},
			wantErr: true,
		},
	}

	for _, test := range listTests {
		test.mock()

		repoUsecase := repositoryUsecase{repositoryRepository: repoMock}

		res, err := repoUsecase.GetRepositoryList(test.args)

		if (err != nil) != test.wantErr {
			t.Errorf("GetRepositoryList() got error : %s", err)
		}

		if test.wantErr {
			assert.Equal(t, test.want, res)
			assert.Error(t, err.Cause())
		}

		if !test.wantErr {
			assert.Equal(t, test.want, res)
		}
	}
}

func TestAddRepository(t *testing.T) {
	repoMock := new(mocks.IRepositoryRepository)
	trxMock := new(mocks.ITrxRepository)

	listTests := []struct {
		name    string
		mock    func()
		args    model.AddRepositoryRequest
		want    interface{}
		wantErr bool
	}{
		{
			name: "ok",
			mock: func() {
				args := model.AddRepositoryRequest{
					Name: "JQuery",
					Url:  "github.com/jquery/jquery",
				}
				w := model.AddRepositoryResponse{
					Id:       3,
					Name:     "JQuery",
					Url:      "github.com/jquery/jquery",
					IsActive: true,
				}
				tx := model.Trx{
					DB: &sqlx.DB{},
					Tx: &sqlx.Tx{},
				}

				repoMock.On("AddRepository", nil, args).Return(w, nil).Once()
				trxMock.On("Create", mock.Anything).Return(&tx, nil).Once()
			},
			args: model.AddRepositoryRequest{
				Name: "JQuery",
				Url:  "github.com/jquery/jquery",
			},
			want: model.AddRepositoryResponse{
				Id:       3,
				Name:     "JQuery",
				Url:      "github.com/jquery/jquery",
				IsActive: true,
			},
			wantErr: false,
		},
		/*{
			name: "error",
			mock: func() {
				e := errors.New("error")
				w := serror.NewFromErrorc(e, "[usecase][AddRepository] add new repository")

				repoMock.On("AddRepository", mock.Anything).
					Return(model.AddRepositoryResponse{}, w).Once()
			},
			want:    model.AddRepositoryResponse{},
			wantErr: true,
		},*/
	}

	for _, test := range listTests {
		test.mock()

		repoUsecase := repositoryUsecase{
			repositoryRepository: repoMock,
			trxRepository:        trxMock,
		}

		res, err := repoUsecase.AddRepository(test.args)

		if (err != nil) != test.wantErr {
			t.Errorf("AddRepository() got error : %s", err)
		}

		if test.wantErr {
			assert.Equal(t, test.want, res)
			assert.Error(t, err.Cause())
		}

		if !test.wantErr {
			assert.Equal(t, test.want, res)
			assert.Equal(t, nil, err)
		}
	}
}
