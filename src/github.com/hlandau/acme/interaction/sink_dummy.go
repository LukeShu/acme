package interaction

type dummySink struct{}

func (dummySink) Close() error {
	return nil
}

func (dummySink) SetProgress(n, ofM int) {
}

func (dummySink) SetStatusLine(status string) {
}
