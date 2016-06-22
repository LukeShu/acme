package acmetool_want

import (
	acmetool_reconcile "github.com/hlandau/acme/cmd/acmetool/acmetool-reconcile"

	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
)

func Main(ctx acmetool.Ctx, reconcile bool, want []string) {
	cmdWant(ctx, want)
	if reconcile {
		acmetool_reconcile.Main(ctx)
	}
}

func cmdWant(ctx acmetool.Ctx, want []string) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

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
	ctx.Logger.Fatale(err, "add target")
}
