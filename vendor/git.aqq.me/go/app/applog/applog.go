package applog

import (
	"fmt"

	"git.aqq.me/go/app/event"
	"go.uber.org/zap"
)

const errPref = "applog"

var logger *Logger

func init() {
	event.Init.AddHandler(initLogger)
	event.Reload.AddHandler(reloadLogger)
	event.Stop.AddHandler(destroyLogger)
}

// GetLogger returns zap logger.
func GetLogger() *zap.Logger {
	if logger == nil {
		panic(fmt.Errorf("%s must be initialized first", errPref))
	}

	return logger.Logger
}

func initLogger() error {
	if logger != nil {
		return nil
	}

	var err error
	logger, err = newLogger()

	if err != nil {
		return err
	}

	return nil
}

func reloadLogger() error {
	var err error
	logger, err = newLogger()

	if err != nil {
		return err
	}

	return nil
}

func destroyLogger() error {
	if logger != nil {
		logger.Close()
		logger = nil
	}

	return nil
}
