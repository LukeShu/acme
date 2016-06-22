package acmetool_reconcile

import (
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
	"github.com/hlandau/xlog"
)

func Main(log xlog.Logger, stateDirName string) {
	s, err := storage.NewFDB(stateDirName)
	log.Fatale(err, "storage")

	err = storageops.Reconcile(s)
	log.Fatale(err, "reconcile")
}
