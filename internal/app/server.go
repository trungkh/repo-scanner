package config

import (
	"repo-scanner/internal/utils/serror"

	"github.com/gin-gonic/gin"
)

func (c *Config) InitServer() serror.SError {

	gin.SetMode(gin.ReleaseMode)
	c.Server = gin.New()

	return nil
}
