package langserver

import (
	"errors"
)

// New returns an error that formats as the given text.
func newClientNotReadyError() error {
	return errors.New("session has not yet been started")
}
