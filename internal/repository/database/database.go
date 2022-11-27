package database

import (
	"errors"
	"fmt"
	"time"

	"repo-scanner/internal/constants"
	"repo-scanner/internal/utils/utint"
	"repo-scanner/internal/utils/utstring"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var availableDriver = []string{"postgres", "mysql"}

type DB struct {
	*sqlx.DB
}

func NewConnection(driver string, connectionString string, connLifeTime int64) (*DB, error) {
	if utstring.ArrContains(availableDriver, driver) == false {
		return nil, errors.New("driver not available")
	}

	db, err := sqlx.Connect(driver, connectionString)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("db: failed connect to database %+v", err))
	}

	db.SetConnMaxLifetime(time.Minute * time.Duration(utint.StringToInt(utstring.Env(constants.DBConnLifetime, "15"), 15)))
	db.SetMaxIdleConns(int(utint.StringToInt(utstring.Env(constants.DBConnMaxIdle, "5"), 5)))
	db.SetMaxOpenConns(int(utint.StringToInt(utstring.Env(constants.DBConnMaxOpen, "0"), 0)))
	return &DB{db}, nil
}

func NewPostgeConnection(connectionString string, connLifeTime int64) (*DB, error) {
	return NewConnection("postgres", connectionString, connLifeTime)
}
