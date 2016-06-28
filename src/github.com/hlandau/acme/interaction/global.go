package interaction

import "github.com/hlandau/xlog"

var log, Log = xlog.New("acme.interactor")

var Stdio Interactor = &stdioInteraction{}

func PrintStderrMessage(title, body string) {
	printStderrMessage(title, body)
}
