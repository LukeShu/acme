package acmetool_import_key

import (
	"io/ioutil"

	"github.com/hlandau/acme/acmeapi/acmeutils"
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
)

func Main(ctx acmetool.Ctx, filename string) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	err = importKey(s, filename)
	ctx.Logger.Fatale(err, "import key")
}

func importKey(s storage.Store, filename string) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	pk, err := acmeutils.LoadPrivateKey(b)
	if err != nil {
		return err
	}

	_, err = s.ImportKey(pk)
	return err
}
