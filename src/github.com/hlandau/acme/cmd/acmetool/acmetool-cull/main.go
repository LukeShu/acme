package acmetool_cull

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
)

func Main(ctx acmetool.Ctx, simulate bool) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	err = storageops.Cull(s, simulate)
	ctx.Logger.Fatale(err, "cull")
}
