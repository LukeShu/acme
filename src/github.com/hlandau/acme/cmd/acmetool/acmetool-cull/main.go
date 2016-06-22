package acmetool_cull

import (
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
	"github.com/hlandau/xlog"
)

func Main(log xlog.Logger, stateDirName string, simulate bool) {
	s, err := storage.NewFDB(stateDirName)
	log.Fatale(err, "storage")

	err = storageops.Cull(s, simulate)
	log.Fatale(err, "cull")
}
