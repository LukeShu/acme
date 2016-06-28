package interaction

import "fmt"

type CannedInteraction struct{
	responses map[string]*Response
}

var _ Interactor = CannedInteraction{}

func NewCannedInteraction() CannedInteraction {
	return CannedInteraction{responses: map[string]*Response{}}
}

func (r CannedInteraction) Status(c *StatusInfo) (StatusSink, error) {
	return nil, fmt.Errorf("not supported")
}

func (r CannedInteraction) Prompt(c *Challenge) (*Response, error) {
	if c.UniqueID == "" {
		return nil, fmt.Errorf("cannot auto-respond to a challenge without a unique ID")
	}

	res := r.responses[c.UniqueID]
	if res == nil {
		return nil, fmt.Errorf("unknown unique ID, cannot respond: %#v", c.UniqueID)
	}

	return res, nil
}

// Configures a canned response for the given interaction UniqueID.
func (r CannedInteraction) SetResponse(uniqueID string, res *Response) {
	res.Noninteractive = true
	r.responses[uniqueID] = res
}
