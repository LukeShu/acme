package acmetool_import_jwk_account

import (
	"io/ioutil"
	"os"

	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	jose "gopkg.in/square/go-jose.v1"
)

func Main(ctx acmetool.Ctx, argURL, argPath string) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	f, err := os.Open(argPath)
	ctx.Logger.Fatale(err, "cannot open private key file")
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	ctx.Logger.Fatale(err, "cannot read file")

	k := jose.JsonWebKey{}
	err = k.UnmarshalJSON(b)
	ctx.Logger.Fatale(err, "cannot unmarshal key")

	_, err = s.ImportAccount(argURL, k.Key)
	ctx.Logger.Fatale(err, "cannot import account key")
}
