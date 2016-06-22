package acmetool_revoke

import (
	"os"

	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
	"github.com/hlandau/xlog"
)

func Main(log xlog.Logger, stateDirName, certSpec string) {
	f, _ := os.Open(certSpec)
	//var fi os.FileInfo
	if f != nil {
		defer f.Close()
		//var err error
		//fi, err = f.Stat()
		//log.Panice(err)
	}
	//u, _ := url.Parse(certSpec)

	switch {
	//case f != nil && !fi.IsDir(): // is a file path

	//case f != nil && fi.IsDir(): // is a directory path
	//  f, _ = os.Open(filepath.Join(certSpec, "cert"))

	//case u != nil && u.IsAbs() && acmeapi.ValidURL(certSpec): // is an URL

	case storage.IsWellFormattedCertificateOrKeyID(certSpec):
		// key or certificate ID
		revokeByCertificateID(log, stateDirName, certSpec)

	default:
		log.Fatalf("don't understand argument, must be a certificate or key ID: %q", certSpec)
	}
}

func revokeByCertificateID(log xlog.Logger, stateDirName string, certID string) {
	s, err := storage.NewFDB(stateDirName)
	log.Fatale(err, "storage")

	err = storageops.RevokeByCertificateOrKeyID(s, certID)
	log.Fatale(err, "revoke")

	err = storageops.Reconcile(s)
	log.Fatale(err, "reconcile")
}
