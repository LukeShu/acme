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
	"github.com/hlandau/acme/acmetool"
	"github.com/hlandau/acme/interaction"
	"github.com/hlandau/degoutils/xlogconfig"
	"github.com/hlandau/xlog"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/hlandau/easyconfig.v1/adaptflag"
	yaml "gopkg.in/yaml.v2"
)

var log, Log = xlog.New("acmetool")

func main() {
	app := &acmetool.App{
		CommandLine: kingpin.New("acmetool", helpText),
		Commands:    map[string]func(acmetool.Ctx){},
	}

	stateFlag := app.CommandLine.Flag("state", "Path to the state directory (env: ACME_STATE_DIR)").
		Default(acmetool.DefaultStateDir).
		Envar("ACME_STATE_DIR").
		PlaceHolder(acmetool.DefaultStateDir).
		String()

	hooksFlag := app.CommandLine.Flag("hooks", "Path to the notification hooks directory (env: ACME_HOOKS_DIR)").
		Default(acmetool.DefaultHooksDir).
		Envar("ACME_HOOKS_DIR").
		PlaceHolder(acmetool.DefaultHooksDir).
		String()

	batchFlag := app.CommandLine.Flag("batch", "Do not attempt interaction; useful for cron jobs. (acmetool can still obtain responses from a response file, if one was provided.)").
		Bool()

	stdioFlag := app.CommandLine.Flag("stdio", "Don't attempt to use console dialogs; fall back to stdio prompts").Bool()

	responseFileFlag := app.CommandLine.Flag("response-file", "Read dialog responses from the given file (default: $ACME_STATE_DIR/conf/responses)").ExistingFile()

	app.CommandLine.Author("Hugo Landau")

	acmetool_reconcile.Register(app)
	acmetool_cull.Register(app)
	acmetool_status.Register(app)
	acmetool_want.Register(app)
	acmetool_unwant.Register(app)
	acmetool_quickstart.Register(app)
	acmetool_redirector.Register(app)
	acmetool_test_notify.Register(app)
	acmetool_import_jwk_account.Register(app)
	acmetool_import_key.Register(app)
	acmetool_import_le.Register(app)
	acmetool_revoke.Register(app)
	acmetool_account_thumbprint.Register(app)

	syscall.Umask(0) // make sure webroot files can be world-readable

	adaptflag.Adapt()
	cmd, err := app.CommandLine.Parse(os.Args[1:])
	if err != nil {
		app.CommandLine.Fatalf("%s, try --help", err)
	}

	*stateFlag, err = filepath.Abs(*stateFlag)
	log.Fatale(err, "state directory path")
	*hooksFlag, err = filepath.Abs(*hooksFlag)
	log.Fatale(err, "hooks directory path")

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

	app.Commands[cmd](acmetool.Ctx{
		Logger:   log,
		StateDir: *stateFlag,
		HooksDir: *hooksFlag,
		Batch:    *batchFlag,
	})
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
