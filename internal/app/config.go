package config

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"repo-scanner/internal/constants"
	"repo-scanner/internal/model"
	"repo-scanner/internal/repository/database"
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/sqlq"
	"repo-scanner/internal/utils/utint"
	"repo-scanner/internal/utils/utstring"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Hostname string
	Server   *gin.Engine
	DB       *database.DB
	Service  *model.Service
	Query    sqlq.SQLQuery
}

func NewApp() Config {
	var config Config

	config.Hostname, _ = os.Hostname()
	config.Service = &model.Service{
		Key:     utstring.Env(constants.AppKey, constants.DefaultAppKey),
		Name:    utstring.Env(constants.AppName, constants.DefaultAppName),
		Version: utstring.Env(constants.AppVersion, "v1.0.0"),
		Host:    utstring.Env(constants.AppHost, "127.0.0.1"),
		Port:    int(utint.StringToInt(utstring.Env(constants.AppPort, utstring.IntToString(constants.DefaultAppPort)), int64(constants.DefaultAppPort))),
	}

	return config
}

func (c *Config) InitEnv() serror.SError {
	if err := godotenv.Load(); err != nil {
		return serror.NewFromError(err)
	}
	return nil
}

func Catch(serr serror.SError) {
	if serr != nil {
		serr.Panic()
	}
}

func (c *Config) Start() (errx serror.SError) {
	ch := make(chan bool)
	go func() {
		log.Info("Running at PORT: ", c.Service.Port)
		err := c.Server.Run(":" + utstring.IntToString(c.Service.Port))
		if err != nil {
			errx = serror.NewFromErrorc(err, "Cannot starting server")
		}
	}()

	<-ch
	return nil
}

func (ox *Config) Stop() {
	//do nothing
}

func most(errx serror.SError) {
	if errx != nil {
		log.Panic(errx)
	}
}
