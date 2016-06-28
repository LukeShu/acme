package acmetool

import (
	"github.com/hlandau/acme/interaction"
	"github.com/hlandau/xlog"
)

type Ctx struct {
	Logger   xlog.Logger
	StateDir string
	HooksDir string
	Interaction interaction.MaybeCannedInteraction
}
