package acmetool_reconcile

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
)

func Main(ctx acmetool.Ctx) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	err = storageops.Reconcile(s)
	ctx.Logger.Fatale(err, "reconcile")
}
