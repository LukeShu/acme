package acmetool_import_jwk_account

import (
	"io/ioutil"
	"os"

	"github.com/hlandau/acme/storage"
	"github.com/hlandau/xlog"
	jose "gopkg.in/square/go-jose.v1"
)

func Main(log xlog.Logger, stateDirName, argURL, argPath string) {
	s, err := storage.NewFDB(stateDirName)
	log.Fatale(err, "storage")

	f, err := os.Open(argPath)
	log.Fatale(err, "cannot open private key file")
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	log.Fatale(err, "cannot read file")

	k := jose.JsonWebKey{}
	err = k.UnmarshalJSON(b)
	log.Fatale(err, "cannot unmarshal key")

	_, err = s.ImportAccount(argURL, k.Key)
	log.Fatale(err, "cannot import account key")
}
