package acmetool_test_notify

import (
	"github.com/hlandau/acme/hooks"
	"github.com/hlandau/xlog"
)

func Main(log xlog.Logger, stateDirName, hooksDirName string, hostnames []string) {
	err := hooks.NotifyLiveUpdated(hooksDirName, stateDirName, hostnames)
	log.Errore(err, "notify")
}
