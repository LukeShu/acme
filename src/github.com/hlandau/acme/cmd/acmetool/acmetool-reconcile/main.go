package acmetool_reconcile

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
)

const reconcileHelp = `Reconcile ACME state, idempotently requesting and renewing certificates to satisfy configured targets.

This is the default command.`

func Register(app *acmetool.App) {
	app.CommandLine.Command("reconcile", reconcileHelp).Default()
	app.Commands["reconcile"] = Main
}

func Main(ctx acmetool.Ctx) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	err = storageops.Reconcile(s, ctx.Interaction)
	ctx.Logger.Fatale(err, "reconcile")
}
