package acmetool_main

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
)

func Main(ctx acmetool.Ctx, unwant []string) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	for _, hn := range unwant {
		err = storageops.RemoveTargetHostname(s, hn)
		ctx.Logger.Fatale(err, "remove target hostname ", hn)
	}
}
