package acmetool_redirector

import (
	"github.com/hlandau/acme/redirector"
	"github.com/hlandau/acme/responder"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/xlog"
	service "gopkg.in/hlandau/service.v2"
)

func Main(log xlog.Logger, stateDirName, rpath, gid string) {
	if rpath == "" {
		// redirector process is internet-facing and must never touch private keys
		storage.Neuter()
		rpath = determineWebroot(log, stateDirName)
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

func determineWebroot(log xlog.Logger, stateDirName string) string {
	s, err := storage.NewFDB(stateDirName)
	log.Fatale(err, "storage")

	webrootPaths := s.DefaultTarget().Request.Challenge.WebrootPaths
	if len(webrootPaths) > 0 {
		return webrootPaths[0]
	}

	return responder.StandardWebrootPath
}
