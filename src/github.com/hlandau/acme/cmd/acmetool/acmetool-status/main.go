package acmetool_status

import (
	"bytes"
	"fmt"

	"github.com/hlandau/acme/acmeapi/acmeutils"
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/acme/storageops"
)

func Main(ctx acmetool.Ctx) {
	s, err := storage.NewFDB(ctx.StateDir)
	ctx.Logger.Fatale(err, "storage")

	info := statusString(ctx, s)
	ctx.Logger.Fatale(err, "status")

	fmt.Print(info)
}

func statusString(ctx acmetool.Ctx, s storage.Store) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Settings:\n")
	fmt.Fprintf(&buf, "  ACME_STATE_DIR: %s\n", s.Path())
	fmt.Fprintf(&buf, "  ACME_HOOKS_DIR: %s\n", ctx.HooksDir)
	fmt.Fprintf(&buf, "  Default directory URL: %s\n", s.DefaultTarget().Request.Provider)
	fmt.Fprintf(&buf, "  Preferred key type: %v\n", &s.DefaultTarget().Request.Key)
	fmt.Fprintf(&buf, "  Additional webroots:\n")
	for _, wr := range s.DefaultTarget().Request.Challenge.WebrootPaths {
		fmt.Fprintf(&buf, "    %s\n", wr)
	}

	fmt.Fprintf(&buf, "\nAvailable accounts:\n")
	s.VisitAccounts(func(a *storage.Account) error {
		fmt.Fprintf(&buf, "  %v\n", a)
		thumbprint, _ := acmeutils.Base64Thumbprint(a.PrivateKey)
		fmt.Fprintf(&buf, "    thumbprint: %s\n", thumbprint)
		return nil
	})

	fmt.Fprintf(&buf, "\n")
	s.VisitTargets(func(t *storage.Target) error {
		fmt.Fprintf(&buf, "%v\n", t)

		c, err := storageops.FindBestCertificateSatisfying(s, t)
		if err != nil {
			fmt.Fprintf(&buf, "  error: %v\n", err)
			return nil // continue
		}

		renewStr := ""
		if storageops.CertificateNeedsRenewing(c) {
			renewStr = " needs-renewing"
		}

		fmt.Fprintf(&buf, "  best: %v%s\n", c, renewStr)
		return nil
	})

	if storageops.HaveUncachedCertificates(s) {
		fmt.Fprintf(&buf, "\nThere are uncached certificates.\n")
	}

	return buf.String()
}
