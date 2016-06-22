package acmetool_test_notify

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/hooks"
)

func Main(ctx acmetool.Ctx, hostnames []string) {
	err := hooks.NotifyLiveUpdated(ctx.HooksDir, ctx.StateDir, hostnames)
	ctx.Logger.Errore(err, "notify")
}
