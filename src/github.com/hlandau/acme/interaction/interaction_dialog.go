package interaction

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

type dialogInteraction struct{
	cmdName string
	cmdType string
}

func NewDialogInteraction() (Interactor, error) {
	cmdName, cmdType := findDialogCommand()
	if cmdName == "" {
		return nil, fmt.Errorf("cannot find whiptail or dialog binary in path")
	}
	di := &dialogInteraction{
		cmdName: cmdName,
		cmdType: cmdType,
	}
	return di, nil
}

type dialogStatusSink struct {
	closeChan  chan struct{}
	closeOnce  sync.Once
	closedChan chan struct{}
	updateChan chan struct{}
	pipeW      *os.File
	infoMutex  sync.Mutex
	statusLine string
	progress   int
	cmd        *exec.Cmd
}

func (ss *dialogStatusSink) Close() error {
	ss.closeOnce.Do(func() {
		close(ss.closeChan)
	})
	<-ss.closedChan
	return nil
}

func (ss *dialogStatusSink) SetProgress(n, ofM int) {
	ss.infoMutex.Lock()
	defer ss.infoMutex.Unlock()
	ss.progress = int((float64(n) / float64(ofM)) * 100)
	ss.notify()
}

func (ss *dialogStatusSink) SetStatusLine(status string) {
	ss.infoMutex.Lock()
	defer ss.infoMutex.Unlock()
	ss.statusLine = status
	ss.notify()
}

func (ss *dialogStatusSink) notify() {
	select {
	case ss.updateChan <- struct{}{}:
	default:
	}
}

func (ss *dialogStatusSink) loop() {
A:
	for {
		select {
		case <-ss.closeChan:
			break A
		case <-ss.updateChan:
			ss.infoMutex.Lock()
			statusLine := ss.statusLine
			progress := ss.progress
			ss.infoMutex.Unlock()

			fmt.Fprintf(ss.pipeW, "XXX\n%d\n%s\nXXX\n", progress, statusLine)
		}
	}

	ss.pipeW.Close()
	ss.cmd.Wait()
	close(ss.closedChan)
}

func (di *dialogInteraction) Status(c *StatusInfo) (StatusSink, error) {

	width := "78"
	height := fmt.Sprintf("%d", strings.Count(c.StatusLine, "\n")+5)

	var opts []string
	if c.Title != "" {
		opts = append(opts, "--backtitle", "ACME", "--title", c.Title)
	}

	opts = append(opts, "--gauge", c.StatusLine, height, width, "0")

	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	defer pipeR.Close()

	cmd := exec.Command(di.cmdName, opts...)
	cmd.Stdin = pipeR
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		pipeW.Close()
		return nil, err
	}

	ss := &dialogStatusSink{
		closeChan:  make(chan struct{}),
		closedChan: make(chan struct{}),
		updateChan: make(chan struct{}, 10),
		pipeW:      pipeW,
		cmd:        cmd,
	}

	go ss.loop()
	return ss, nil
}

func (di *dialogInteraction) Prompt(c *Challenge) (*Response, error) {
	width := "78"
	height := "49"
	var yesLabelArg, noLabelArg, noTagsArg string
	switch di.cmdType {
	case "dialog":
		yesLabelArg = "--yes-label"
		noLabelArg = "--no-label"
		noTagsArg = "--no-tags"
	case "whiptail":
		yesLabelArg = "--yes-button"
		noLabelArg = "--no-button"
		noTagsArg = "--notags"
	default:
		panic("invalid value of cmdType")
	}

	var opts []string
	if c.Title != "" {
		opts = append(opts, "--backtitle", "ACME", "--title", c.Title)
	}

	var err error
	var pipeR *os.File
	var pipeW *os.File

	switch c.ResponseType {
	case RTAcknowledge:
		opts = append(opts, "--msgbox", c.Body, height, width)
	case RTYesNo:
		yesLabel := c.YesLabel
		if yesLabel == "" {
			yesLabel = "Yes"
		}
		noLabel := c.NoLabel
		if noLabel == "" {
			noLabel = "No"
		}
		opts = append(opts, yesLabelArg, yesLabel, noLabelArg, noLabel, "--yesno", c.Body, height, width)
	case RTLineString:
		pipeR, pipeW, err = os.Pipe()
		if err != nil {
			return nil, err
		}

		defer pipeR.Close()
		defer pipeW.Close()
		opts = append(opts, "--output-fd", "3", "--inputbox", c.Body, height, width)
	case RTSelect:
		pipeR, pipeW, err = os.Pipe()
		if err != nil {
			return nil, err
		}

		defer pipeR.Close()
		defer pipeW.Close()
		opts = append(opts, "--output-fd", "3", noTagsArg, "--menu", c.Body, height, width, "5")
		for _, o := range c.Options {
			opts = append(opts, o.Value, o.Title)
		}
	}

	cmd := exec.Command(di.cmdName, opts...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if pipeW != nil {
		cmd.ExtraFiles = append(cmd.ExtraFiles, pipeW)
	}

	rc, xerr, err := runCommand(cmd)
	if err != nil {
		return nil, err
	}

	// If we get error code >1 (particularly 255) the dialog command probably
	// doesn't support some option we pass it. Return an error, which should make
	// us fall back to stdio.
	if rc > 1 {
		return nil, xerr
	}

	res := &Response{}
	if pipeW != nil {
		pipeW.Close()
	}

	switch c.ResponseType {
	case RTLineString, RTSelect:
		b, err := ioutil.ReadAll(pipeR)
		if err != nil {
			return nil, err
		}

		res.Value = string(b)
		fallthrough
	case RTYesNo, RTAcknowledge:
		if rc != 0 && rc != 1 {
			return nil, xerr
		}
		res.Cancelled = (rc == 1)
	}

	return res, nil
}

func findDialogCommand() (string, string) {
	// not using whiptail for now, see #18
	for _, s := range []string{"dialog"} {
		p, err := exec.LookPath(s)
		if err == nil {
			return p, s
		}
	}

	return "", ""
}

func runCommand(cmd *exec.Cmd) (int, error, error) {
	err := cmd.Run()
	if err == nil {
		return 0, nil, nil
	}

	if e, ok := err.(*exec.ExitError); ok {
		if ws, ok := e.Sys().(syscall.WaitStatus); ok {
			return ws.ExitStatus(), err, nil
		}
	}

	return 255, err, err
}
