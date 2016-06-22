package acmetool_want

import (
	acmetool_reconcile "github.com/hlandau/acme/cmd/acmetool/acmetool-reconcile"

	"github.com/hlandau/acme/storage"
	"github.com/hlandau/xlog"
)

func Main(log xlog.Logger, stateDirName string, reconcile bool, want []string) {
	cmdWant(log, stateDirName, want)
	if reconcile {
		acmetool_reconcile.Main(log, stateDirName)
	}
}

func cmdWant(log xlog.Logger, stateDirName string, want []string) {
	s, err := storage.NewFDB(stateDirName)
	log.Fatale(err, "storage")

	alreadyExists := false
	s.VisitTargets(func(t *storage.Target) error {
		nm := map[string]struct{}{}
		for _, n := range t.Satisfy.Names {
			nm[n] = struct{}{}
		}

		for _, w := range want {
			if _, ok := nm[w]; !ok {
				return nil
			}
		}

		alreadyExists = true
		return nil
	})

	if alreadyExists {
		return
	}

	tgt := storage.Target{
		Satisfy: storage.TargetSatisfy{
			Names: want,
		},
	}

	err = s.SaveTarget(&tgt)
	log.Fatale(err, "add target")
}
