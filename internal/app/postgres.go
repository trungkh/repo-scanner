package config

import (
	"fmt"
	"time"

	"repo-scanner/internal/constants"
	"repo-scanner/internal/repository/database"
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utstring"
	"repo-scanner/internal/utils/uttime"

	"github.com/gearintellix/u2"
)

func (c *Config) InitPosgres() serror.SError {
	sqlConn := utstring.Env(constants.DBConnStr, `
        host=__host__
        user=__user__
        password=__pwd__
		port=__port__
        dbname=__name__
        sslmode=__sslMode__
        application_name=__appKey__
    `)
	sqlConn = u2.Binding(sqlConn, map[string]string{
		"host":    utstring.Env(constants.DBHost, "localhost"),
		"port":    utstring.Env(constants.DBPort, "5432"),
		"user":    utstring.Env(constants.DBUser, "postgres"),
		"pwd":     utstring.Env(constants.DBPwd, "root"),
		"name":    utstring.Env(constants.DBName, "postgres"),
		"sslMode": utstring.Env(constants.DBSSLMode, "disabled"),
		"appKey":  utstring.Env(constants.AppKey, ""),
		"appName": utstring.Env(constants.AppName, ""),
		"date":    uttime.ToString(uttime.DefaultDateFormat, time.Now()),
	})

	db, err := database.NewPostgeConnection(sqlConn, 15)
	if err != nil {
		note := fmt.Sprintf("failed connect to database %s:%s/%s user:%s",
			utstring.Env(constants.DBHost, "localhost"),
			utstring.Env(constants.DBPort, "5432"),
			utstring.Env(constants.DBName, "postgres"),
			utstring.Env(constants.DBUser, "postgres"))
		return serror.NewFromErrorc(err, note)
	}
	c.DB = db

	return nil
}
