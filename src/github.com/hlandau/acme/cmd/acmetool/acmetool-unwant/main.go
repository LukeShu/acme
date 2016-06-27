package acmetool_main

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
)

func Register(app *acmetool.App) {
	cmd := app.CommandLine.Command("unwant", "Modify targets to remove any mentions of the given hostnames")
	unwant := cmd.Arg("hostname", "hostnames which should be removed from all target files").Required().Strings()
	app.Commands["unwant"] = func(ctx acmetool.Ctx) { Main(ctx, *unwant) }
}

func Main(ctx acmetool.Ctx, unwant []string) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	for _, hn := range unwant {
		err = storageops.RemoveTargetHostname(s, hn)
		ctx.Logger.Fatale(err, "remove target hostname ", hn)
	}
}
