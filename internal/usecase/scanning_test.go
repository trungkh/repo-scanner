package usecase

import (
	"repo-scanner/internal/mocks"
	"repo-scanner/internal/model"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetScanningList(t *testing.T) {
	repoMock := new(mocks.IRepositoryRepository)
	scanMock := new(mocks.IScanningRepository)

	listTests := []struct {
		name    string
		mock    func()
		args    model.ScanningListRequest
		want    []model.ScanningListResponse
		wantErr bool
	}{
		{
			name: "ok",
			mock: func() {
				w := []model.ScanningListResponse{
					{
						Id:         3,
						Name:       "JQuery",
						Url:        "github.com/jquery/jquery",
						Findings:   types.JSONText([]byte(`{}`)),
						Status:     "queued",
						QueuedAt:   time.Time{},
						ScanningAt: &time.Time{},
						FinishedAt: &time.Time{},
					},
				}

				scanMock.On("GetScanningList", mock.Anything).Return(w, nil).Once()
			},
			args: model.ScanningListRequest{
				Limit:  1,
				Page:   1,
				Sort:   "desc",
				Status: "all",
			},
			want: []model.ScanningListResponse{
				{
					Id:         3,
					Name:       "JQuery",
					Url:        "github.com/jquery/jquery",
					Findings:   types.JSONText([]byte(`{}`)),
					Status:     "queued",
					QueuedAt:   time.Time{},
					ScanningAt: &time.Time{},
					FinishedAt: &time.Time{},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range listTests {
		test.mock()

		scanUsecase := scanningUsecase{
			repositoryRepository: repoMock,
			scanningRepository:   scanMock,
		}

		res, err := scanUsecase.GetScanningList(test.args)

		if (err != nil) != test.wantErr {
			t.Errorf("GetScanningList() got error : %s", err)
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

func TestAddNewScanning(t *testing.T) {
	repoMock := new(mocks.IRepositoryRepository)
	scanMock := new(mocks.IScanningRepository)
	trxMock := new(mocks.ITrxRepository)
	txMock := new(mocks.ITrx)

	listTests := []struct {
		name    string
		mock    func()
		args    int64
		want    model.ScanningResponse
		wantErr bool
	}{
		{
			name: "ok",
			mock: func() {
				w := model.ScanningResponse{
					Id:         10,
					RepoId:     3,
					Findings:   types.JSONText([]byte(`{}`)),
					Status:     "queued",
					QueuedAt:   time.Time{},
					ScanningAt: &time.Time{},
					FinishedAt: &time.Time{},
				}
				r := model.Repository{
					IsActive: true,
				}
				tx := model.Trx{
					DB: &sqlx.DB{},
				}

				repoMock.On("GetRepositoryById", mock.Anything).Return(&r, nil).Once()
				scanMock.On("AddNewScanning", &tx, int64(3)).Return(w, nil).Once()
				trxMock.On("Create", mock.Anything).Return(&tx, nil).Once()
				txMock.On("Admit", mock.Anything).Return(nil).Once()
			},
			args: 3,
			want: model.ScanningResponse{
				Id:         10,
				RepoId:     3,
				Findings:   types.JSONText([]byte(`{}`)),
				Status:     "queued",
				QueuedAt:   time.Time{},
				ScanningAt: &time.Time{},
				FinishedAt: &time.Time{},
			},
			wantErr: false,
		},
	}

	for _, test := range listTests {
		test.mock()

		scanUsecase := scanningUsecase{
			repositoryRepository: repoMock,
			scanningRepository:   scanMock,
			trxRepository:        trxMock,
		}

		res, err := scanUsecase.AddNewScanning(test.args)

		if (err != nil) != test.wantErr {
			t.Errorf("AddNewScanning() got error : %s", err)
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

func TestStartScanningInQueue(t *testing.T) {
	repoMock := new(mocks.IRepositoryRepository)
	scanMock := new(mocks.IScanningRepository)
	grabMock := new(mocks.IGrabScanner)

	listTests := []struct {
		name    string
		mock    func()
		wantErr bool
	}{
		{
			name: "ok",
			mock: func() {
				scanMock.On("GetScanningList", mock.Anything).Return([]model.ScanningListResponse{}, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, test := range listTests {
		test.mock()

		scanUsecase := scanningUsecase{
			repositoryRepository: repoMock,
			scanningRepository:   scanMock,
			grabScanner:          grabMock,
		}

		err := scanUsecase.StartScanningInQueue()

		if (err != nil) != test.wantErr {
			t.Errorf("StartScanningInQueue() got error : %s", err)
		}

		if test.wantErr {
			assert.Error(t, err.Cause())
		}
	}
}
