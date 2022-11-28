package postgres

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"repo-scanner/internal"
	"repo-scanner/internal/constants"
	"repo-scanner/internal/model"
	"repo-scanner/internal/repository/database"
	"repo-scanner/internal/repository/postgres/queries"
	"repo-scanner/internal/utils/sqlq"
	"repo-scanner/internal/utils/uttime"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

func NewScanningMock() (internal.IScanningRepository, *sql.DB, sqlmock.Sqlmock) {
	log.SetOutput(ioutil.Discard)
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	opts := sqlq.BuilderOption{
		Driver: sqlq.DriverPostgreSQL,
	}

	builder := sqlq.NewBuilder(opts)

	sqlxDb := sqlx.NewDb(db, "sqlmock")
	postDB := &database.DB{sqlxDb}

	trxRepo := NewTrxRepository(nil)

	repo := NewScanningRepository(postDB, builder, trxRepo)
	return repo, db, mock
}

func TestGetScanningList(t *testing.T) {
	repo, db, mock := NewScanningMock()
	defer func() {
		db.Close()
	}()

	wayback := time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)
	expectQuery := fmt.Sprintf(queries.GetScanningList, "desc")

	tests := []struct {
		name        string
		repo        internal.IScanningRepository
		mock        func()
		requestBody model.ScanningListRequest
		want        []model.ScanningListResponse
		wantErr     bool
	}{
		{
			name: "OK",
			repo: repo,
			mock: func() {

				rows := sqlmock.NewRows([]string{
					"scanning_id",
					"repository_name",
					"repository_url",
					"findings",
					"scanning_status",
					"queued_at",
					"scanning_at",
					"finished_at",
				}).AddRow(
					10,
					"JQuery",
					"github.com/jquery/jquery",
					types.JSONText([]byte(`{}`)),
					"queued",
					currentTime,
					currentTime,
					currentTime,
				)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(expectQuery)).WithArgs(
					"all",
					1,
					0,
				)
				expectedQuery.WillReturnRows(rows)
			},
			requestBody: model.ScanningListRequest{
				Limit:  1,
				Page:   1,
				Sort:   "desc",
				Status: "all",
			},
			want: []model.ScanningListResponse{
				{
					Id:         10,
					Name:       "JQuery",
					Url:        "github.com/jquery/jquery",
					Findings:   types.JSONText([]byte(`{}`)),
					Status:     "queued",
					QueuedAt:   currentTime,
					ScanningAt: &currentTime,
					FinishedAt: &currentTime,
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		test.mock()
		got, err := repo.GetScanningList(test.requestBody)
		if (err != nil) != test.wantErr {
			t.Errorf("GetScanningList() error '%s'", err)
			return
		}

		if err == nil {
			assert.Equal(t, reflect.DeepEqual(got, test.want), true)
		}
	}
}

func TestAddNewScanning(t *testing.T) {
	repo, db, mock := NewScanningMock()
	defer func() {
		db.Close()
	}()

	wayback := time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)

	tests := []struct {
		name        string
		repo        internal.IScanningRepository
		mock        func()
		requestBody int64
		want        model.ScanningResponse
		wantErr     bool
	}{
		{
			name: "OK",
			repo: repo,
			mock: func() {
				rows := sqlmock.NewRows([]string{
					"scanning_id",
					"repository_id",
					"findings",
					"scanning_status",
					"queued_at",
					"scanning_at",
					"finished_at",
				}).AddRow(
					10,
					3,
					types.JSONText([]byte(`{}`)),
					"queued",
					currentTime,
					currentTime,
					currentTime,
				)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(queries.InsertNewScanning)).WithArgs(
					3,
					currentTime,
					"Anonymous",
					currentTime,
					"Anonymous",
					currentTime,
				)
				expectedQuery.WillReturnRows(rows)
			},
			requestBody: 3,
			want: model.ScanningResponse{
				Id:         10,
				RepoId:     3,
				Findings:   types.JSONText([]byte(`{}`)),
				Status:     "queued",
				QueuedAt:   currentTime,
				ScanningAt: &currentTime,
				FinishedAt: &currentTime,
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		test.mock()
		got, err := repo.AddNewScanning(nil, test.requestBody)
		if (err != nil) != test.wantErr {
			t.Errorf("AddNewScanning() error '%s'", err)
			return
		}

		if err == nil {
			assert.Equal(t, reflect.DeepEqual(got, test.want), true)
		}
	}
}

func TestEditScanningStatusById(t *testing.T) {
	repo, db, mock := NewScanningMock()
	defer func() {
		db.Close()
	}()

	wayback := time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)

	tests := []struct {
		name          string
		repo          internal.IScanningRepository
		mock          func()
		reqScanningId int64
		reqStatus     string
		reqFindings   types.JSONText
		want          model.ScanningResponse
		wantErr       bool
	}{
		{
			name: "OK",
			repo: repo,
			mock: func() {
				rows := sqlmock.NewRows([]string{
					"scanning_id",
					"repository_id",
					"findings",
					"scanning_status",
					"queued_at",
					"scanning_at",
					"finished_at",
				}).AddRow(
					10,
					3,
					types.JSONText([]byte(`{}`)),
					"in_progress",
					currentTime,
					currentTime,
					currentTime,
				)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(queries.UpdateScanningInProgressById)).WithArgs(
					10,
					"in_progress",
					"Automated",
					currentTime,
				)
				expectedQuery.WillReturnRows(rows)
			},
			reqScanningId: 10,
			reqStatus:     "in_progress",
			reqFindings:   types.JSONText([]byte(`{}`)),
			want: model.ScanningResponse{
				Id:         10,
				RepoId:     3,
				Findings:   types.JSONText([]byte(`{}`)),
				Status:     "in_progress",
				QueuedAt:   currentTime,
				ScanningAt: &currentTime,
				FinishedAt: &currentTime,
			},
			wantErr: false,
		},
		{
			name: "OK",
			repo: repo,
			mock: func() {
				rows := sqlmock.NewRows([]string{
					"scanning_id",
					"repository_id",
					"findings",
					"scanning_status",
					"queued_at",
					"scanning_at",
					"finished_at",
				}).AddRow(
					10,
					3,
					types.JSONText([]byte(`{"result":"success"}`)),
					"success",
					currentTime,
					currentTime,
					currentTime,
				)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(queries.UpdateScanningFinishedById)).WithArgs(
					10,
					"success",
					types.JSONText([]byte(`{"result":"success"}`)),
					"Automated",
					currentTime,
				)
				expectedQuery.WillReturnRows(rows)
			},
			reqScanningId: 10,
			reqStatus:     "success",
			reqFindings:   types.JSONText([]byte(`{"result":"success"}`)),
			want: model.ScanningResponse{
				Id:         10,
				RepoId:     3,
				Findings:   types.JSONText([]byte(`{"result":"success"}`)),
				Status:     "success",
				QueuedAt:   currentTime,
				ScanningAt: &currentTime,
				FinishedAt: &currentTime,
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		test.mock()
		got, err := repo.EditScanningStatusById(nil, test.reqScanningId, test.reqStatus, test.reqFindings)
		if (err != nil) != test.wantErr {
			t.Errorf("EditScanningStatusById() error '%s'", err)
			return
		}

		if err == nil {
			assert.Equal(t, reflect.DeepEqual(got, test.want), true)
		}
	}
}
