package acmetool_revoke

import (
	"os"

	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
)

func Main(ctx acmetool.Ctx, certSpec string) {
	f, _ := os.Open(certSpec)
	//var fi os.FileInfo
	if f != nil {
		defer f.Close()
		//var err error
		//fi, err = f.Stat()
		//ctx.Logger.Panice(err)
	}
	//u, _ := url.Parse(certSpec)

	switch {
	//case f != nil && !fi.IsDir(): // is a file path

	//case f != nil && fi.IsDir(): // is a directory path
	//  f, _ = os.Open(filepath.Join(certSpec, "cert"))

	//case u != nil && u.IsAbs() && acmeapi.ValidURL(certSpec): // is an URL

	case storage.IsWellFormattedCertificateOrKeyID(certSpec):
		// key or certificate ID
		revokeByCertificateID(ctx, certSpec)

	default:
		ctx.Logger.Fatalf("don't understand argument, must be a certificate or key ID: %q", certSpec)
	}
}

func revokeByCertificateID(ctx acmetool.Ctx, certID string) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	err = storageops.RevokeByCertificateOrKeyID(s, certID)
	ctx.Logger.Fatale(err, "revoke")

	err = storageops.Reconcile(s)
	ctx.Logger.Fatale(err, "reconcile")
}
