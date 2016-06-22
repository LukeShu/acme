package acmetool_redirector

import (
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/redirector"
	"github.com/hlandau/acme/storage"
	service "gopkg.in/hlandau/service.v2"
)

func Main(ctx acmetool.Ctx, rpath, gid string) {
	if rpath == "" {
		// redirector process is internet-facing and must never touch private keys
		storage.Neuter()
		rpath = determineWebroot(ctx)
	}

	service.Main(&service.Info{
		Name:          "acmetool",
		Description:   "acmetool HTTP redirector",
		DefaultChroot: rpath,
		NewFunc: func() (service.Runnable, error) {
			return redirector.New(redirector.Config{
				Bind:          ":80",
				ChallengePath: rpath,
				ChallengeGID:  gid,
			})
		},
	})
}

func determineWebroot(ctx acmetool.Ctx) string {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	webrootPaths := s.DefaultTarget().Request.Challenge.WebrootPaths
	if len(webrootPaths) > 0 {
		return webrootPaths[0]
	}

	return acmetool.DefaultWebRootDir
}
