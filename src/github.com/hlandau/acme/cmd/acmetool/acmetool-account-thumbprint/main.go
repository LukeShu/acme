package acmetool_account_thumbprint

import (
	"fmt"

	"github.com/hlandau/acme/acmeapi/acmeutils"
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
)

func Main(ctx acmetool.Ctx) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	s.VisitAccounts(func(a *storage.Account) error {
		thumbprint, _ := acmeutils.Base64Thumbprint(a.PrivateKey)
		fmt.Printf("%s\t%s\n", thumbprint, a.ID())
		return nil
	})
}
