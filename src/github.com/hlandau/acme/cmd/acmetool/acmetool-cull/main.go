package acmetool_cull

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
)

func Register(app *acmetool.App) {
	cmd := app.CommandLine.Command("cull", "Delete expired, unused certificates")
	simulate := cmd.Flag("simulate", "Show which certificates would be deleted without deleting any").Short('n').Bool()
	app.Commands["cull"] = func(ctx acmetool.Ctx) { Main(ctx, *simulate) }
}

func Main(ctx acmetool.Ctx, simulate bool) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	err = storageops.Cull(s, simulate)
	ctx.Logger.Fatale(err, "cull")
}
