package acmetool_import_key

import (
	"io/ioutil"

	"github.com/hlandau/acme/acmeapi/acmeutils"
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
)

func Register(app *acmetool.App) {
	cmd := app.CommandLine.Command("import-key", "Import a certificate private key")
	filename := cmd.Arg("private-key-file", "Path to PEM-encoded private key").Required().ExistingFile()
	app.Commands["import-key"] = func(ctx acmetool.Ctx) { Main(ctx, *filename) }
}

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
