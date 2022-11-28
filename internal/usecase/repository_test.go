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
	txMock := new(mocks.ITrx)

	listTests := []struct {
		name    string
		mock    func()
		args    model.AddRepositoryRequest
		want    model.AddRepositoryResponse
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
				}

				repoMock.On("AddRepository", &tx, args).Return(w, nil).Once()
				trxMock.On("Create", mock.Anything).Return(&tx, nil).Once()
				txMock.On("Admit", mock.Anything).Return(nil).Once()
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

func TestEditRepository(t *testing.T) {
	repoMock := new(mocks.IRepositoryRepository)
	trxMock := new(mocks.ITrxRepository)
	txMock := new(mocks.ITrx)

	repoName := "JQuery"

	listTests := []struct {
		name    string
		mock    func()
		args    model.EditRepositoryRequest
		want    model.EditRepositoryResponse
		wantErr bool
	}{
		{
			name: "ok",
			mock: func() {
				args := model.EditRepositoryRequest{
					Id:       3,
					Name:     &repoName,
					Url:      nil,
					IsActive: nil,
				}
				w := model.EditRepositoryResponse{
					Id:       3,
					Name:     "JQuery",
					Url:      "github.com/jquery/jquery",
					IsActive: true,
				}
				o := model.Repository{
					Id:       3,
					Name:     "New JQuery",
					Url:      "github.com/jquery/jquery",
					IsActive: true,
				}
				tx := model.Trx{
					DB: &sqlx.DB{},
				}

				repoMock.On("GetRepositoryById", mock.Anything).Return(&o, nil).Once()
				repoMock.On("EditRepository", &tx, args).Return(w, nil).Once()
				trxMock.On("Create", mock.Anything).Return(&tx, nil).Once()
				txMock.On("Admit", mock.Anything).Return(nil).Once()
			},
			args: model.EditRepositoryRequest{
				Id:       3,
				Name:     &repoName,
				Url:      nil,
				IsActive: nil,
			},
			want: model.EditRepositoryResponse{
				Id:       3,
				Name:     "JQuery",
				Url:      "github.com/jquery/jquery",
				IsActive: true,
			},
			wantErr: false,
		},
	}

	for _, test := range listTests {
		test.mock()

		repoUsecase := repositoryUsecase{
			repositoryRepository: repoMock,
			trxRepository:        trxMock,
		}

		res, err := repoUsecase.EditRepository(test.args)

		if (err != nil) != test.wantErr {
			t.Errorf("EditRepository() got error : %s", err)
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
