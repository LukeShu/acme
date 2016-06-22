package acmetool_account_thumbprint

import (
	"fmt"

	"github.com/hlandau/acme/acmeapi/acmeutils"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/xlog"
)

func Main(log xlog.Logger, stateDirName string) {
	s, err := storage.NewFDB(stateDirName)
	log.Fatale(err, "storage")

	s.VisitAccounts(func(a *storage.Account) error {
		thumbprint, _ := acmeutils.Base64Thumbprint(a.PrivateKey)
		fmt.Printf("%s\t%s\n", thumbprint, a.ID())
		return nil
	})
}
