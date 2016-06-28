package interaction

import (
	"fmt"
)

type MaybeCannedInteraction struct {
	Canned CannedInteraction
	Fresh Interactor
}

var _ Interactor = MaybeCannedInteraction{}

func (ai MaybeCannedInteraction) Prompt(c *Challenge) (*Response, error) {
	r, err := ai.Canned.Prompt(c)
	if err == nil || c.Implicit {
		return r, err
	}
	log.Infoe(err, "interaction auto-responder couldn't give a canned response")

	if ai.Fresh == nil {
		return nil, fmt.Errorf("cannot prompt the user: currently non-interactive")
	}

	return ai.Fresh.Prompt(c)
}

func (ai MaybeCannedInteraction) Status(info *StatusInfo) (StatusSink, error) {
	if ai.Fresh == nil {
		return dummySink{}, nil
	} else {
		return ai.Fresh.Status(info)
	}
}
