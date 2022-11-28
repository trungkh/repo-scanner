package postgres

import (
	"database/sql"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"testing"
	"time"

	"repo-scanner/internal"
	"repo-scanner/internal/constants"
	"repo-scanner/internal/model"
	"repo-scanner/internal/repository/database"
	"repo-scanner/internal/repository/postgres/queries"
	"repo-scanner/internal/utils/sqlq"
	"repo-scanner/internal/utils/uttime"

	"bou.ke/monkey"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func NewRepositoryMock() (internal.IRepositoryRepository, *sql.DB, sqlmock.Sqlmock) {
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

	repo := NewRepositoryRepository(postDB, builder, trxRepo)
	return repo, db, mock
}

func TestGetRepositoryList(t *testing.T) {
	repo, db, mock := NewRepositoryMock()
	defer func() {
		db.Close()
	}()

	tests := []struct {
		name        string
		repo        internal.IRepositoryRepository
		mock        func()
		requestBody model.RepositoryListRequest
		want        []model.RepositoryListResponse
		wantErr     bool
	}{
		{
			name: "OK",
			repo: repo,
			mock: func() {
				rows := sqlmock.NewRows([]string{
					"repository_id",
					"repository_name",
					"repository_url",
					"is_active",
				}).AddRow(
					3,
					"JQuery",
					"github.com/jquery/jquery",
					true)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(queries.GetRepositoryList)).WithArgs(1, 0)
				expectedQuery.WillReturnRows(rows)
			},
			requestBody: model.RepositoryListRequest{
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
	}

	for _, test := range tests {
		test.mock()
		got, err := repo.GetRepositoryList(test.requestBody)
		if (err != nil) != test.wantErr {
			t.Errorf("GetRepositoryList() error '%s'", err)
			return
		}

		if err == nil {
			assert.Equal(t, reflect.DeepEqual(got, test.want), true)
		}
	}
}

func TestGetRepositoryById(t *testing.T) {
	repo, db, mock := NewRepositoryMock()
	defer func() {
		db.Close()
	}()

	tests := []struct {
		name        string
		repo        internal.IRepositoryRepository
		mock        func()
		requestBody int64
		want        *model.Repository
		wantErr     bool
	}{
		{
			name: "OK",
			repo: repo,
			mock: func() {
				rows := sqlmock.NewRows([]string{
					"repository_id",
					"repository_name",
					"repository_url",
					"is_active",
					"created_by",
					"created_at",
					"modified_by",
					"modified_at",
					"deleted_by",
					"deleted_at",
				}).AddRow(
					3,
					"JQuery",
					"github.com/jquery/jquery",
					true,
					"Anonymous",
					time.Time{},
					"Anonymous",
					time.Time{},
					nil,
					nil,
				)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(queries.GetRepositoryById)).WithArgs(3)
				expectedQuery.WillReturnRows(rows)
			},
			requestBody: 3,
			want: &model.Repository{
				Id:         3,
				Name:       "JQuery",
				Url:        "github.com/jquery/jquery",
				IsActive:   true,
				CreatedBy:  "Anonymous",
				CreatedAt:  time.Time{},
				ModifiedBy: "Anonymous",
				ModifiedAt: time.Time{},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		test.mock()
		got, err := repo.GetRepositoryById(test.requestBody)
		if (err != nil) != test.wantErr {
			t.Errorf("GetRepositoryById() error '%s'", err)
			return
		}

		if err == nil {
			assert.Equal(t, reflect.DeepEqual(got, test.want), true)
		}
	}
}

func TestAddRepository(t *testing.T) {
	repo, db, mock := NewRepositoryMock()
	defer func() {
		db.Close()
	}()

	wayback := time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	tests := []struct {
		name        string
		repo        internal.IRepositoryRepository
		mock        func()
		requestBody model.AddRepositoryRequest
		want        model.AddRepositoryResponse
		wantErr     bool
	}{
		{
			name: "OK",
			repo: repo,
			mock: func() {
				rows := sqlmock.NewRows([]string{
					"repository_id",
					"repository_name",
					"repository_url",
					"is_active",
				}).AddRow(
					3,
					"JQuery",
					"github.com/jquery/jquery",
					true)

				currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(queries.InsertNewRepository)).WithArgs(
					"JQuery",
					"github.com/jquery/jquery",
					"Anonymous",
					currentTime,
					"Anonymous",
					currentTime,
				)
				expectedQuery.WillReturnRows(rows)
			},
			requestBody: model.AddRepositoryRequest{
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

	for _, test := range tests {
		test.mock()
		got, err := repo.AddRepository(nil, test.requestBody)
		if (err != nil) != test.wantErr {
			t.Errorf("AddRepository() error '%s'", err)
			return
		}

		if err == nil {
			assert.Equal(t, reflect.DeepEqual(got, test.want), true)
		}
	}
}

func TestEditRepository(t *testing.T) {
	repo, db, mock := NewRepositoryMock()
	defer func() {
		db.Close()
	}()

	wayback := time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	tests := []struct {
		name        string
		repo        internal.IRepositoryRepository
		mock        func()
		requestBody model.EditRepositoryRequest
		want        model.EditRepositoryResponse
		wantErr     bool
	}{
		{
			name: "OK",
			repo: repo,
			mock: func() {
				rows := sqlmock.NewRows([]string{
					"repository_id",
					"repository_name",
					"repository_url",
					"is_active",
				}).AddRow(
					3,
					"JQuery",
					"github.com/jquery/jquery",
					true)

				currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(queries.EditRepository)).WithArgs(
					3,
					false, nil,
					false, nil,
					false, nil,
					"Anonymous",
					currentTime,
				)
				expectedQuery.WillReturnRows(rows)
			},
			requestBody: model.EditRepositoryRequest{
				Id:       3,
				Name:     nil,
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

	for _, test := range tests {
		test.mock()
		got, err := repo.EditRepository(nil, test.requestBody)
		if (err != nil) != test.wantErr {
			t.Errorf("EditRepository() error '%s'", err)
			return
		}

		if err == nil {
			assert.Equal(t, reflect.DeepEqual(got, test.want), true)
		}
	}
}

/*func TestDeleteRepository(t *testing.T) {
	repo, db, mock := NewRepositoryMock()
	defer func() {
		db.Close()
	}()

	wayback := time.Date(1974, time.May, 19, 1, 2, 3, 4, time.UTC)
	patch := monkey.Patch(time.Now, func() time.Time { return wayback })
	defer patch.Unpatch()

	tests := []struct {
		name        string
		repo        internal.IRepositoryRepository
		mock        func()
		requestBody int64
		wantErr     bool
	}{
		{
			name: "OK",
			repo: repo,
			mock: func() {
				currentTime, _ := uttime.NowWithTimezone(constants.DefaultTimezone)
				expectedQuery := mock.ExpectQuery(regexp.QuoteMeta(queries.DeleteRepository)).WithArgs(
					3,
					"Anonymous",
					currentTime,
				)
				expectedQuery.RowsWillBeClosed()
			},
			requestBody: 3,
			wantErr:     false,
		},
	}

	for _, test := range tests {
		test.mock()
		err := repo.DeleteRepository(nil, test.requestBody)
		if (err != nil) != test.wantErr {
			t.Errorf("EditRepository() error '%s'", err)
			return
		}
	}
}*/
