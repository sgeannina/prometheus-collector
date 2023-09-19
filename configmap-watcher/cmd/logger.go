package cmd

import (
	"go.goms.io/aks/rp/toolkit/log"
)

const (
	configmapWatcherSource = "ConfigmapWatcherLog"
)

// newAddonTokenReconcilerLogger creates a logger supposed to be used by addon token reconciler
func newConfigmapWatcherLogger() *log.Logger {
	logrus := log.GetGlobalLogger()

	return &log.Logger{
		TraceLogger: logrus.WithField(log.SourceFieldName, configmapWatcherSource),
	}
}
