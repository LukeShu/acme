// acmetool, an automated certificate acquisition tool for ACME servers.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	acmetool_account_thumbprint "github.com/hlandau/acme/cmd/acmetool/acmetool-account-thumbprint"
	acmetool_cull "github.com/hlandau/acme/cmd/acmetool/acmetool-cull"
	acmetool_import_jwk_account "github.com/hlandau/acme/cmd/acmetool/acmetool-import-jwk-account"
	acmetool_import_key "github.com/hlandau/acme/cmd/acmetool/acmetool-import-key"
	acmetool_import_le "github.com/hlandau/acme/cmd/acmetool/acmetool-import-le"
	acmetool_quickstart "github.com/hlandau/acme/cmd/acmetool/acmetool-quickstart"
	acmetool_reconcile "github.com/hlandau/acme/cmd/acmetool/acmetool-reconcile"
	acmetool_redirector "github.com/hlandau/acme/cmd/acmetool/acmetool-redirector"
	acmetool_revoke "github.com/hlandau/acme/cmd/acmetool/acmetool-revoke"
	acmetool_status "github.com/hlandau/acme/cmd/acmetool/acmetool-status"
	acmetool_test_notify "github.com/hlandau/acme/cmd/acmetool/acmetool-test-notify"
	acmetool_unwant "github.com/hlandau/acme/cmd/acmetool/acmetool-unwant"
	acmetool_want "github.com/hlandau/acme/cmd/acmetool/acmetool-want"

	"github.com/hlandau/acme/acmeapi"
	"github.com/hlandau/acme/hooks"
	"github.com/hlandau/acme/interaction"
	"github.com/hlandau/acme/storage"
	"github.com/hlandau/degoutils/xlogconfig"
	"github.com/hlandau/xlog"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/hlandau/easyconfig.v1/adaptflag"
	yaml "gopkg.in/yaml.v2"
)

var log, Log = xlog.New("acmetool")

var (
	stateFlag = kingpin.Flag("state", "Path to the state directory (env: ACME_STATE_DIR)").
			Default(storage.RecommendedPath).
			Envar("ACME_STATE_DIR").
			PlaceHolder(storage.RecommendedPath).
			String()

	hooksFlag = kingpin.Flag("hooks", "Path to the notification hooks directory (env: ACME_HOOKS_DIR)").
			Default(hooks.RecommendedPath).
			Envar("ACME_HOOKS_DIR").
			PlaceHolder(hooks.RecommendedPath).
			String()

	batchFlag = kingpin.Flag("batch", "Do not attempt interaction; useful for cron jobs. (acmetool can still obtain responses from a response file, if one was provided.)").
			Bool()

	stdioFlag = kingpin.Flag("stdio", "Don't attempt to use console dialogs; fall back to stdio prompts").Bool()

	responseFileFlag = kingpin.Flag("response-file", "Read dialog responses from the given file (default: $ACME_STATE_DIR/conf/responses)").ExistingFile()

	reconcileCmd = kingpin.Command("reconcile", reconcileHelp).Default()

	cullCmd          = kingpin.Command("cull", "Delete expired, unused certificates")
	cullSimulateFlag = cullCmd.Flag("simulate", "Show which certificates would be deleted without deleting any").Short('n').Bool()

	statusCmd = kingpin.Command("status", "Show active configuration")

	wantCmd       = kingpin.Command("want", "Add a target with one or more hostnames")
	wantReconcile = wantCmd.Flag("reconcile", "Specify --no-reconcile to skip reconcile after adding target").Default("1").Bool()
	wantArg       = wantCmd.Arg("hostname", "hostnames for which a certificate should be obtained").Required().Strings()

	unwantCmd = kingpin.Command("unwant", "Modify targets to remove any mentions of the given hostnames")
	unwantArg = unwantCmd.Arg("hostname", "hostnames which should be removed from all target files").Required().Strings()

	quickstartCmd = kingpin.Command("quickstart", "Interactively ask some getting started questions (recommended)")
	expertFlag    = quickstartCmd.Flag("expert", "Ask more questions in quickstart wizard").Bool()

	redirectorCmd      = kingpin.Command("redirector", "HTTP to HTTPS redirector with challenge response support")
	redirectorPathFlag = redirectorCmd.Flag("path", "Path to serve challenge files from").String()
	redirectorGIDFlag  = redirectorCmd.Flag("challenge-gid", "GID to chgrp the challenge path to (optional)").String()

	testNotifyCmd = kingpin.Command("test-notify", "Test-execute notification hooks as though given hostnames were updated")
	testNotifyArg = testNotifyCmd.Arg("hostname", "hostnames which have been updated").Strings()

	importJWKAccountCmd = kingpin.Command("import-jwk-account", "Import a JWK account key")
	importJWKURLArg     = importJWKAccountCmd.Arg("provider-url", "Provider URL (e.g. https://acme-v01.api.letsencrypt.org/directory)").Required().String()
	importJWKPathArg    = importJWKAccountCmd.Arg("private-key-file", "Path to private_key.json").Required().ExistingFile()

	importKeyCmd = kingpin.Command("import-key", "Import a certificate private key")
	importKeyArg = importKeyCmd.Arg("private-key-file", "Path to PEM-encoded private key").Required().ExistingFile()

	importLECmd = kingpin.Command("import-le", "Import a Let's Encrypt client state directory")
	importLEArg = importLECmd.Arg("le-state-path", "Path to Let's Encrypt state directory").Default("/etc/letsencrypt").ExistingDir()

	// Arguments we should probably support for revocation:
	//   A certificate ID
	//   A key ID
	//   A path to a PEM-encoded certificate - TODO
	//   A path to a PEM-encoded private key (revoke all known certificates with that key) - TODO
	//   A path to a certificate directory - TODO
	//   A path to a key directory - TODO
	//   A certificate URL - TODO
	revokeCmd = kingpin.Command("revoke", "Revoke a certificate")
	revokeArg = revokeCmd.Arg("certificate-id-or-path", "Certificate ID to revoke").String()

	accountThumbprintCmd = kingpin.Command("account-thumbprint", "Prints account thumbprints")
)

const reconcileHelp = `Reconcile ACME state, idempotently requesting and renewing certificates to satisfy configured targets.

This is the default command.`

func main() {
	syscall.Umask(0) // make sure webroot files can be world-readable

	adaptflag.Adapt()
	cmd := kingpin.Parse()

	var err error
	*stateFlag, err = filepath.Abs(*stateFlag)
	log.Fatale(err, "state directory path")
	*hooksFlag, err = filepath.Abs(*hooksFlag)
	log.Fatale(err, "hooks directory path")

	hooks.DefaultPath = *hooksFlag
	acmeapi.UserAgent = "acmetool"
	xlogconfig.Init()

	if *batchFlag {
		interaction.NonInteractive = true
	}

	if *stdioFlag {
		interaction.NoDialog = true
	}

	if *responseFileFlag == "" {
		p := filepath.Join(*stateFlag, "conf/responses")
		if _, err := os.Stat(p); err == nil {
			*responseFileFlag = p
		}
	}

	if *responseFileFlag != "" {
		err := loadResponseFile(*responseFileFlag)
		log.Errore(err, "cannot load response file, continuing anyway")
	}

	switch cmd {
	case "reconcile":
		acmetool_reconcile.Main(log, *stateFlag)
	case "cull":
		acmetool_cull.Main(log, *stateFlag, *cullSimulateFlag)
	case "status":
		acmetool_status.Main(log, *stateFlag)
	case "account-thumbprint":
		acmetool_account_thumbprint.Main(log, *stateFlag)
	case "want":
		acmetool_want.Main(log, *stateFlag, *wantReconcile, *wantArg)
	case "unwant":
		acmetool_unwant.Main(log, *stateFlag, *unwantArg)
	case "quickstart":
		acmetool_quickstart.Main(log, *stateFlag, *hooksFlag, *batchFlag, *expertFlag)
	case "redirector":
		acmetool_redirector.Main(log, *stateFlag, *redirectorPathFlag, *redirectorGIDFlag)
	case "test-notify":
		acmetool_test_notify.Main(log, *stateFlag, *hooksFlag, *testNotifyArg)
	case "import-key":
		acmetool_import_key.Main(log, *stateFlag, *importKeyArg)
	case "import-jwk-account":
		acmetool_import_jwk_account.Main(log, *stateFlag, *importJWKURLArg, *importJWKPathArg)
	case "import-le":
		acmetool_import_le.Main(log, *stateFlag, *importLEArg)
	case "revoke":
		acmetool_revoke.Main(log, *stateFlag, *revokeArg)
	}
}

// YAML response file loading.

func loadResponseFile(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	m := map[string]interface{}{}
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	for k, v := range m {
		r, err := parseResponse(v)
		if err != nil {
			log.Errore(err, "response for ", k, " invalid")
			continue
		}
		interaction.SetResponse(k, r)
	}

	return nil
}

func parseResponse(v interface{}) (*interaction.Response, error) {
	switch x := v.(type) {
	case string:
		return &interaction.Response{
			Value: x,
		}, nil
	case int:
		return &interaction.Response{
			Value: fmt.Sprintf("%d", x),
		}, nil
	case bool:
		return &interaction.Response{
			Cancelled: !x,
		}, nil
	default:
		return nil, fmt.Errorf("unknown response value")
	}
}
