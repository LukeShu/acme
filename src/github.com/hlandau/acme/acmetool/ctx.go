package acmetool

import "github.com/hlandau/xlog"

type Ctx struct {
	Logger   xlog.Logger
	StateDir string
	HooksDir string
	Batch    bool
}
