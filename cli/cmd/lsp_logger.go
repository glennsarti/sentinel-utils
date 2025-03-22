package cmd

import (
	"io"

	"github.com/glennsarti/sentinel-utils/cli/ui"
)

var _ io.Writer = &lspLogWriter{}

// StdOut/StdErr Writer
type lspLogWriter struct {
	localUi   ui.Ui
	useStdErr bool
}

func (w lspLogWriter) Write(p []byte) (n int, err error) {
	if w.useStdErr {
		w.localUi.Error(string(p))
	} else {
		w.localUi.Output(string(p))
	}
	return len(p), nil
}

func newLogWriter(localUi ui.Ui, useStdErr bool) *lspLogWriter {
	return &lspLogWriter{
		localUi:   localUi,
		useStdErr: useStdErr,
	}
}
