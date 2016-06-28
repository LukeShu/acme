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

type interactionFlags struct {
	Mode string
	ResponseFile string
}

func commandLineParser(ctx *acmetool.Ctx, iFlags *interactionFlags) *acmetool.App {
	app := &acmetool.App{
		CommandLine: kingpin.New("acmetool", helpText),
		Commands:    map[string]func(acmetool.Ctx){},
	}
	app.CommandLine.Author("Hugo Landau")

	app.CommandLine.Flag("state", "Path to the state directory (env: ACME_STATE_DIR)").
		Default(acmetool.DefaultStateDir).
		Envar("ACME_STATE_DIR").
		PlaceHolder(acmetool.DefaultStateDir).
		StringVar(&ctx.StateDir)

	app.CommandLine.Flag("hooks", "Path to the notification hooks directory (env: ACME_HOOKS_DIR)").
		Default(acmetool.DefaultHooksDir).
		Envar("ACME_HOOKS_DIR").
		PlaceHolder(acmetool.DefaultHooksDir).
		StringVar(&ctx.HooksDir)

	app.CommandLine.Flag("interaction","Set the interaction mode."+
		"\"batch\" disables interaction (useful for cron jobs, may require a response file to be provided); "+
		"\"dialog\" uses the `dialog` program to create text user interfaces; "+
		"\"stdio\" uses plain terminal prompts; "+
		"\"auto\" uses dialog if avialable, or falls back to stdio").
		Default("auto").
		PlaceHolder("<batch|dialog|stdio|auto>").
		EnumVar(&iFlags.Mode, "batch", "dialog", "stdio", "auto")

	app.CommandLine.Flag("response-file", "Read dialog responses from the given file (default: $ACME_STATE_DIR/conf/responses)").
		ExistingFileVar(&iFlags.ResponseFile)

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

	adaptflag.AdaptWithFunc(func(info adaptflag.Info) {
		dpn := adaptflag.DottedPath(info.Path)
		if len(dpn) > 0 {
			dpn += "."
		}
		dpn += info.Name
		app.CommandLine.Flag(dpn, info.Usage).SetValue(info.Value)
	})

	return app
}

func main() {
	var ctx acmetool.Ctx
	var iFlags interactionFlags
	app := commandLineParser(&ctx, &iFlags)
	ctx.Logger, _ = xlog.New("acmetool")
	ctx.Interaction = interaction.MaybeCannedInteraction{
		Canned: interaction.NewCannedInteraction(),
		Fresh: nil,
	}

	syscall.Umask(0) // make sure webroot files can be world-readable
	acmeapi.UserAgent = "acmetool"
	xlogconfig.Init()

	cmd, err := app.CommandLine.Parse(os.Args[1:])
	if err != nil {
		app.CommandLine.Fatalf("%s, try --help", err)
	}

	ctx.StateDir, err = filepath.Abs(ctx.StateDir)
	ctx.Logger.Fatale(err, "state directory path")
	ctx.HooksDir, err = filepath.Abs(ctx.HooksDir)
	ctx.Logger.Fatale(err, "hooks directory path")

	switch iFlags.Mode {
	case "batch":
		ctx.Interaction.Fresh = nil
	case "dialog":
		ctx.Interaction.Fresh, err = interaction.NewDialogInteraction()
		ctx.Logger.Fatale(err, "")
	case "stdio":
		ctx.Interaction.Fresh = interaction.Stdio
	case "auto":
		ctx.Interaction.Fresh, err = interaction.NewDialogInteraction()
		if err != nil {
			ctx.Interaction.Fresh = interaction.Stdio
		}
	default:
		panic("invalid result from kingpin Enum")
	}

	if iFlags.ResponseFile == "" {
		p := filepath.Join(ctx.StateDir, "conf/responses")
		if _, err := os.Stat(p); err == nil {
			iFlags.ResponseFile = p
		}
	}
	if iFlags.ResponseFile != "" {
		err := loadResponseFile(ctx.Logger, iFlags.ResponseFile, &ctx.Interaction.Canned)
		ctx.Logger.Errore(err, "cannot load response file, continuing anyway")
	}

	app.Commands[cmd](ctx)
}

// YAML response file loading.

func loadResponseFile(log xlog.Logger, path string, canned *interaction.CannedInteraction) error {
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
		canned.SetResponse(k, r)
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
