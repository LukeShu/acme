package interaction

type InteractionSaver struct{
	inner Interactor
	responsesReceived map[string]*Response
}

var _ Interactor = InteractionSaver{}

func NewInteractionSaver(inner Interactor) InteractionSaver {
	return InteractionSaver{
		inner: inner,
		responsesReceived: map[string]*Response{},
	}
}

func (is InteractionSaver) Prompt(c *Challenge) (*Response, error) {
	res, err := is.inner.Prompt(c)
	if err == nil && c.UniqueID != "" {
		is.responsesReceived[c.UniqueID] = res
	}

	return res, err
}

func (is InteractionSaver) Status(si *StatusInfo) (StatusSink, error) {
	return is.inner.Status(si)
}

// Returns a map from challenge UniqueIDs to responses received for those
// UniqueIDs. Do not mutate the returned map.
func (is InteractionSaver) ResponsesReceived() map[string]*Response {
	return is.responsesReceived
}
