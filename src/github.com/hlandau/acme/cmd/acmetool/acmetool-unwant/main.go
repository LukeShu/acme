package acmetool_main

import (
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
	"github.com/hlandau/xlog"
)

func Main(log xlog.Logger, stateDirName string, unwant []string) {
	s, err := storage.NewFDB(stateDirName)
	log.Fatale(err, "storage")

	for _, hn := range unwant {
		err = storageops.RemoveTargetHostname(s, hn)
		log.Fatale(err, "remove target hostname ", hn)
	}
}
