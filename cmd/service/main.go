package main

import (
	config "repo-scanner/internal/app"

	log "github.com/sirupsen/logrus"
)

func main() {
	// global log level
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	app := config.NewApp()

	config.Catch(app.InitEnv())
	config.Catch(app.InitPosgres())
	config.Catch(app.InitQuery())
	config.Catch(app.InitServer())
	config.Catch(app.InitService())
	defer app.Stop()

	log.Info("Starting Repository Scanner...")
	config.Catch(app.Start())
}
