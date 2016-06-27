package acmetool_test_notify

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/hooks"
)

func Register(app *acmetool.App) {
	cmd := app.CommandLine.Command("test-notify", "Test-execute notification hooks as though given hostnames were updated")
	hostnames := cmd.Arg("hostname", "hostnames which have been updated").Strings()
	app.Commands["test-notify"] = func(ctx acmetool.Ctx) { Main(ctx, *hostnames) }
}

func Main(ctx acmetool.Ctx, hostnames []string) {
	err := hooks.NotifyLiveUpdated(ctx.HooksDir, ctx.StateDir, hostnames)
	ctx.Logger.Errore(err, "notify")
}
